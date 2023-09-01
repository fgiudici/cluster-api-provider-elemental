/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/rancher-sandbox/cluster-api-provider-elemental/api/v1beta1"
)

const (
	defaultRequeuePeriod = 10 * time.Second
)

// ElementalMachineReconciler reconciles a ElementalMachine object
type ElementalMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElementalMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.ElementalMachine{}).
		Watches(
			&infrastructurev1beta1.ElementalHost{},
			handler.EnqueueRequestsFromMapFunc(r.ElementalHostToElementalMachine),
		).
		Watches(
			&clusterv1.Machine{},
			handler.EnqueueRequestsFromMapFunc(r.MachineToElementalMachine),
		).
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(r.ClusterToElementalMachines),
			// Reconciliation step #5: If the associated Cluster‘s status.infrastructureReady is false, exit the reconciliation
			// Note: This check should not be blocking any further delete reconciliation flows.
			// Note: This check should only be performed after appropriate owner references (if any) are updated.
			builder.WithPredicates(predicates.ClusterUnpausedAndInfrastructureReady(ctrl.LoggerFrom(ctx))),
		).
		Complete(r)
}

func (r *ElementalMachineReconciler) ElementalHostToElementalMachine(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx).WithValues("host", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	logger.Info("Enqueueing ElementalMachine reconciliation from ElementalHost")

	requests := []ctrl.Request{}

	// Verify we are actually handling a ElementalHost object
	host, ok := obj.(*infrastructurev1beta1.ElementalHost)
	if !ok {
		errMsg := fmt.Sprintf("Expected a ElementalHost object, but got %T", obj)
		logger.Error(errors.New(errMsg), errMsg)
		return []ctrl.Request{}
	}

	// Check the ElementalHost was associated to any ElementalMachine
	if host.Status.MachineRef != nil {
		logger.Info("Adding ElementalMachine to reconciliation request", "elementalMachine", fmt.Sprintf("%s/%s", host.Status.MachineRef.Namespace, host.Status.MachineRef.Name))
		name := client.ObjectKey{Namespace: host.Status.MachineRef.Namespace, Name: host.Status.MachineRef.Name}
		requests = append(requests, ctrl.Request{NamespacedName: name})
	}

	return requests
}

func (r *ElementalMachineReconciler) MachineToElementalMachine(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx).WithValues("host", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	logger.Info("Enqueueing ElementalMachine reconciliation from Machine")

	requests := []ctrl.Request{}
	// Verify we are actually handling a Machine object
	machine, ok := obj.(*clusterv1.Machine)
	if !ok {
		errMsg := fmt.Sprintf("Expected a Machine object, but got %T", obj)
		logger.Error(errors.New(errMsg), errMsg)
		return []ctrl.Request{}
	}

	// Check the Machine was associated to any ElementalMachine
	if machine.Spec.InfrastructureRef.Kind == "ElementalMachine" {
		logger.Info("Adding ElementalMachine to reconciliation request", "elementalMachine", fmt.Sprintf("%s/%s", machine.Spec.InfrastructureRef.Namespace, machine.Spec.InfrastructureRef.Name))
		name := client.ObjectKey{Namespace: machine.Spec.InfrastructureRef.Namespace, Name: machine.Spec.InfrastructureRef.Name}
		requests = append(requests, ctrl.Request{NamespacedName: name})
	}

	return requests
}

// ClusterToElementalMachines is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation of ElementalMachines.
func (r *ElementalMachineReconciler) ClusterToElementalMachines(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.FromContext(ctx).WithValues("cluster", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	logger.Info("Enqueueing ElementalMachines reconciliation from Cluster")

	requests := []ctrl.Request{}

	// Verify we are actually handling a Cluster object
	cluster, ok := obj.(*clusterv1.Cluster)
	if !ok {
		errMsg := fmt.Sprintf("Expected a Cluster object, but got %T", obj)
		logger.Error(errors.New(errMsg), errMsg)
		return []ctrl.Request{}
	}

	// Fetch the MachineList associated to this Cluster
	labels := map[string]string{clusterv1.ClusterNameLabel: cluster.Name}
	capiMachineList := &clusterv1.MachineList{}
	if err := r.Client.List(ctx, capiMachineList, client.InNamespace(cluster.Namespace),
		client.MatchingLabels(labels),
	); err != nil {
		logger.Error(err, "failed to list ElementalMachines")
		return []ctrl.Request{}
	}

	// Enqueue related (same NamespacedName of Machines) ElementalMachines for reconciliation
	for i, m := range capiMachineList.Items {
		if m.Spec.InfrastructureRef.Name == "" {
			continue
		}
		name := client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.InfrastructureRef.Name}
		if m.Spec.InfrastructureRef.Namespace != "" {
			name = client.ObjectKey{Namespace: m.Spec.InfrastructureRef.Namespace, Name: m.Spec.InfrastructureRef.Name}
		}
		logger.Info("Adding ElementalMachine to reconciliation request", "elementalMachine", fmt.Sprintf("%s/%s", capiMachineList.Items[i].Namespace, capiMachineList.Items[i].Name))
		requests = append(requests, ctrl.Request{NamespacedName: name})
	}

	return requests
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=elementalmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=elementalmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=elementalmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;machines,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
// For more details about the reconciliation loop, check the official CAPI documentation:
// - https://cluster-api.sigs.k8s.io/developer/providers/machine-infrastructure
func (r *ElementalMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ElementalMachine")

	// Fetch the ElementalMachine
	elementalMachine := &infrastructurev1beta1.ElementalMachine{}
	if err := r.Client.Get(ctx, req.NamespacedName, elementalMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("fetching ElementalMachine: %w", err)
	}

	// Reconciliation step #1: If the resource does not have a Machine owner, exit the reconciliation
	machine, err := util.GetOwnerMachine(ctx, r.Client, elementalMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("getting Machine owner: %w", err)
	}
	if machine == nil {
		logger.Info("ElementalMachine resource has no Machine owner")
		return ctrl.Result{}, nil
	}

	// Reconciliation step #2: If the resource has status.failureReason or status.failureMessage set, exit the reconciliation
	// TODO: status.failureReason and failureMessage not status.implemented yet.

	// Fetch the Cluster
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, elementalMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("fetching Cluster: %w", err)
	}

	// Reconciliation step #3: If the Cluster to which this resource belongs cannot be found, exit the reconciliation
	if cluster == nil {
		logger.Info("ElementalMachine resource is not associated with any Cluster")
		return ctrl.Result{}, nil
	}

	// Create the patch helper.
	patchHelper, err := patch.NewHelper(elementalMachine, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("initializing patch helper: %w", err)
	}
	// Always issue a patch when exiting this function so changes to the
	// resource are patched back to the API server.
	defer func() {
		// Reconciliation step #12: Patch the resource to persist changes
		patchHelper.Patch(ctx, elementalMachine)
	}()

	if elementalMachine.GetDeletionTimestamp() == nil || elementalMachine.GetDeletionTimestamp().IsZero() {
		// The object is not being deleted, so register the finalizer
		if !controllerutil.ContainsFinalizer(elementalMachine, infrastructurev1beta1.FinalizerElementalMachine) {
			// Reconciliation step #4: Add the provider-specific finalizer, if needed
			controllerutil.AddFinalizer(elementalMachine, infrastructurev1beta1.FinalizerElementalMachine)
		}
		// Reconciliation step #5: If the associated Cluster‘s status.infrastructureReady is false, exit the reconciliation
		if !cluster.Status.InfrastructureReady {
			logger.Info("Cluster status.infrastructureReady is false")
			return ctrl.Result{}, nil
		}
		// Reconciliation step #6: If the associated Machine‘s spec.bootstrap.dataSecretName is nil, exit the reconciliation
		if machine.Spec.Bootstrap.DataSecretName == nil {
			logger.Info("Machine spec.bootstrap.dataSecretName is nil")
			return ctrl.Result{}, nil
		}
		// Reconciliation step #7: Reconcile provider-specific machine infrastructure
		if result, err := r.reconcileNormal(ctx, elementalMachine, *machine); err != nil {
			// Reconciliation step #7-1: If they are terminal failures, set status.failureReason and status.failureMessage
			// TODO: Consider implementing status.failureReason and status.failureMessage
			return ctrl.Result{}, fmt.Errorf("reconciling ElementalMachine: %w", err)
		} else {
			return result, nil
		}
	} else {
		// The object is up for deletion
		if controllerutil.ContainsFinalizer(elementalMachine, infrastructurev1beta1.FinalizerElementalMachine) {
			if err := r.reconcileDelete(ctx, elementalMachine); err != nil {
				return ctrl.Result{}, fmt.Errorf("reconciling ElementalMachine deletion: %w", err)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ElementalMachineReconciler) reconcileNormal(ctx context.Context, elementalMachine *infrastructurev1beta1.ElementalMachine, machine clusterv1.Machine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Normal ElementalMachine reconcile")
	// Reconciliation step #7-2: If this is a control plane machine, register the instance with the provider’s control plane load balancer (optional)
	// TODO: Not implemented yet.

	// elementalMachine.Spec.ProviderID is used to mark a link between the ElementalMachine and an ElementalHost
	if elementalMachine.Spec.ProviderID == nil {
		return r.associateElementalHost(ctx, elementalMachine, machine)
	}

	// Reconciliation step #9: Set status.ready to true
	host := &infrastructurev1beta1.ElementalHost{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: elementalMachine.Status.HostRef.Namespace, Name: elementalMachine.Status.HostRef.Name}, host); err != nil {
		return ctrl.Result{}, fmt.Errorf("fetching associated ElementalHost: %w", err)
	}
	if host.Status.Installed && host.Status.Bootstrapped {
		logger.Info("ElementalMachine is ready")
		elementalMachine.Status.Ready = true
	}

	// Reconciliation step #11: Set spec.failureDomain to the provider-specific failure domain the instance is running in (optional)
	// TODO: Not implemented yet.
	return ctrl.Result{}, nil
}

func (r *ElementalMachineReconciler) associateElementalHost(ctx context.Context, elementalMachine *infrastructurev1beta1.ElementalMachine, machine clusterv1.Machine) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Finding a suitable ElementalHost to associate")
	elementalHosts := &infrastructurev1beta1.ElementalHostList{}
	var selector labels.Selector
	var selectorErr error
	// Use the label selector defined in the ElementalMachine, or select any ElementalHost available if no selector has been defined.
	if elementalMachine.Spec.Selector != nil {
		if selector, selectorErr = metav1.LabelSelectorAsSelector(elementalMachine.Spec.Selector); selectorErr != nil {
			return ctrl.Result{}, fmt.Errorf("converting LabelSelector to Selector: %w", selectorErr)
		}
	} else {
		selector = labels.NewSelector()
	}

	if err := r.Client.List(ctx, elementalHosts, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, fmt.Errorf("listing available ElementalHosts: %w", err)
	}

	// Associate the first ElementalHost in the list that is ready to be bootstrapped
	var elementalHostCandidate *infrastructurev1beta1.ElementalHost
	for _, host := range elementalHosts.Items {
		if host.Status.Installed {
			elementalHostCandidate = &host
		}
	}
	if elementalHostCandidate == nil {
		logger.Info("No ElementalHosts available for association. Waiting for new hosts to be provisioned.")
		return ctrl.Result{RequeueAfter: defaultRequeuePeriod}, nil
	}

	logger = logger.WithValues("elementalHost", fmt.Sprintf("%s/%s", elementalHostCandidate.Namespace, elementalHostCandidate.Name))
	logger.Info("Associating ElementalMachine to ElementalHost")

	// Reconciliation step #8: Set spec.providerID to the provider-specific identifier for the provider’s machine instance
	providerID := fmt.Sprintf("elemental://%s/%s", elementalHostCandidate.Namespace, elementalHostCandidate.Name)
	elementalMachine.Spec.ProviderID = &providerID
	elementalMachine.Status.HostRef = &corev1.ObjectReference{
		APIVersion: elementalHostCandidate.APIVersion,
		Kind:       elementalHostCandidate.Kind,
		Namespace:  elementalHostCandidate.Namespace,
		Name:       elementalHostCandidate.Name,
		UID:        elementalHostCandidate.UID,
	}

	// Create the patch helper.
	patchHelper, err := patch.NewHelper(elementalHostCandidate, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("initializing patch helper: %w", err)
	}

	// Link the ElementalMachine to ElementalHost
	elementalHostCandidate.Status.MachineRef = &corev1.ObjectReference{
		APIVersion: elementalMachine.APIVersion,
		Kind:       elementalMachine.Kind,
		Namespace:  elementalMachine.Namespace,
		Name:       elementalMachine.Name,
		UID:        elementalMachine.UID,
	}

	// Link Bootstrap Secret to ElementalHost
	elementalHostCandidate.Spec.BootstrapSecret = &corev1.ObjectReference{
		Kind:      "Secret",
		Namespace: machine.Namespace,
		Name:      *machine.Spec.Bootstrap.DataSecretName,
	}

	// TODO: Decorate the ElementalHost with useful labels, for example the Cluster name, Control Plane endpoint, etc.

	// Reconciliation step #10: Set status.addresses to the provider-specific set of instance addresses
	// TODO: Fetch the addresses from ElementalHost to update the associated ElementalMachine

	// Patch the associated ElementalHost
	if err := patchHelper.Patch(ctx, elementalHostCandidate); err != nil {
		return ctrl.Result{}, fmt.Errorf("patching ElementalHost: %w", err)
	}

	logger.Info("Association successful")

	return ctrl.Result{}, nil
}

func (r *ElementalMachineReconciler) reconcileDelete(ctx context.Context, elementalMachine *infrastructurev1beta1.ElementalMachine) error {
	logger := log.FromContext(ctx)
	logger.Info("Deletion ElementalMachine reconcile")
	controllerutil.RemoveFinalizer(elementalMachine, infrastructurev1beta1.FinalizerElementalMachine)
	return nil
}
