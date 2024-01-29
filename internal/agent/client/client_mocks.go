// /*
// Copyright © 2022 - 2023 SUSE LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rancher-sandbox/cluster-api-provider-elemental/internal/agent/client (interfaces: Client)
//
// Generated by this command:
//
//	mockgen -copyright_file=hack/boilerplate.go.txt -destination=internal/agent/client/client_mocks.go -package=client github.com/rancher-sandbox/cluster-api-provider-elemental/internal/agent/client Client
//
// Package client is a generated GoMock package.
package client

import (
	reflect "reflect"

	config "github.com/rancher-sandbox/cluster-api-provider-elemental/internal/agent/config"
	api "github.com/rancher-sandbox/cluster-api-provider-elemental/internal/api"
	identity "github.com/rancher-sandbox/cluster-api-provider-elemental/internal/identity"
	vfs "github.com/twpayne/go-vfs/v4"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// CreateHost mocks base method.
func (m *MockClient) CreateHost(arg0 api.HostCreateRequest, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateHost", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateHost indicates an expected call of CreateHost.
func (mr *MockClientMockRecorder) CreateHost(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateHost", reflect.TypeOf((*MockClient)(nil).CreateHost), arg0, arg1)
}

// DeleteHost mocks base method.
func (m *MockClient) DeleteHost(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteHost", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteHost indicates an expected call of DeleteHost.
func (mr *MockClientMockRecorder) DeleteHost(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteHost", reflect.TypeOf((*MockClient)(nil).DeleteHost), arg0)
}

// GetBootstrap mocks base method.
func (m *MockClient) GetBootstrap(arg0 string) (*api.BootstrapResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBootstrap", arg0)
	ret0, _ := ret[0].(*api.BootstrapResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBootstrap indicates an expected call of GetBootstrap.
func (mr *MockClientMockRecorder) GetBootstrap(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBootstrap", reflect.TypeOf((*MockClient)(nil).GetBootstrap), arg0)
}

// GetRegistration mocks base method.
func (m *MockClient) GetRegistration(arg0 string) (*api.RegistrationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegistration", arg0)
	ret0, _ := ret[0].(*api.RegistrationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRegistration indicates an expected call of GetRegistration.
func (mr *MockClientMockRecorder) GetRegistration(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegistration", reflect.TypeOf((*MockClient)(nil).GetRegistration), arg0)
}

// Init mocks base method.
func (m *MockClient) Init(arg0 vfs.FS, arg1 identity.Identity, arg2 config.Config) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockClientMockRecorder) Init(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockClient)(nil).Init), arg0, arg1, arg2)
}

// PatchHost mocks base method.
func (m *MockClient) PatchHost(arg0 api.HostPatchRequest, arg1 string) (*api.HostResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PatchHost", arg0, arg1)
	ret0, _ := ret[0].(*api.HostResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PatchHost indicates an expected call of PatchHost.
func (mr *MockClientMockRecorder) PatchHost(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PatchHost", reflect.TypeOf((*MockClient)(nil).PatchHost), arg0, arg1)
}
