package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	infrastructurev1beta3 "github.com/rancher-sandbox/cluster-api-provider-elemental/api/v1beta3"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) GetMachineRegistration(response http.ResponseWriter, request *http.Request) {
	pathVars := mux.Vars(request)
	namespace := pathVars["namespace"]
	registrationName := pathVars["registrationName"]

	// Fetch registration
	registration := &infrastructurev1beta3.ElementalMachineRegistration{}
	if err := s.k8sClient.Get(request.Context(), k8sclient.ObjectKey{Namespace: namespace, Name: registrationName}, registration); err != nil {
		if k8sapierrors.IsNotFound(err) {
			response.WriteHeader(http.StatusNotFound)
			response.Write([]byte(fmt.Sprintf("ElementalMachineRegistration '%s' not found", registrationName)))
		} else {
			s.logger.Error(err, "Could not fetch ElementalMachineRegistration", "namespace", namespace, "registrationName", registrationName)
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(fmt.Sprintf("Could not fetch ElementalMachineRegistration '%s'", registrationName)))
		}
		return
	}

	// Serialize to JSON
	responseBytes, err := json.Marshal(registration)
	if err != nil {
		s.logger.Error(err, "Could not encode response body", "registration", fmt.Sprintf("%+v", registration))
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(fmt.Errorf("Could not encode response body: %w", err).Error()))
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(responseBytes)
}
