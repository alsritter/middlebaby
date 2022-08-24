package apimanager

import (
	"sync"

	"github.com/alsritter/middlebaby/pkg/interact"
)

// MockCaseCenter save all type mock.
type MockCaseCenter interface {
	// AddHttp add something http mock
	AddHttp(uniqID string, mockHttp ...interact.HttpImposter) error
	// AddGlobalHttp add something global http mock
	AddGlobalHttp(mockHttp ...interact.HttpImposter)
	// GetGlobalHttp get something global http mock
	GetGlobalHttp() []interact.HttpImposter
	// GetHttp get something http mock
	GetHttp(uniqID string) ([]interact.HttpImposter, bool)
	// GetAllHttp get all mock, including global mocks
	GetAllHttp() []interact.HttpImposter
	// UnLoadAllHttp unload mock all of a Case
	UnLoadAllHttp()
	// UnLoadAllGlobalHttp unload mock all global of a Case
	UnLoadAllGlobalHttp()
	// UnLoadHttp unload mock specified
	UnLoadHttp(uniqID string)
	// UnloadHttpByIdList unload the mock with Id in httpIdList
	UnloadHttpByIdList(uniqID string, httpIdList []string)

	// AddGRpc add something gRpc mock
	AddGRpc(uniqID string, mockGRpc ...interact.GRpcImposter) error
	// GetGRpc get something gRpc mock
	GetGRpc(uniqID string) ([]interact.GRpcImposter, bool)
	// GetAllGRpc get all gRpc mock
	GetAllGRpc() []interact.GRpcImposter
	// UnLoadGRpc unload specified gRpc mock
	UnLoadGRpc(uniqID string)
	// UnloadGRpcByIdList unload the mock with Id in grpcIdList
	UnloadGRpcByIdList(uniqID string, grpcIdList []string)
	// AddGlobalGRpc add something global grpc mock
	AddGlobalGRpc(mockGRpc ...interact.GRpcImposter)
	// GetGlobalGRpc get something global grpc mock
	GetGlobalGRpc() []interact.GRpcImposter
}

const (
	globalHttpID = "globalHttpID"
	globalGRpcID = "globalGRpcID"
)

// mockCaseCenter storage all mock cases
type mockCaseCenter struct {
	httpMock map[string][]interact.HttpImposter
	gRpcMock map[string][]interact.GRpcImposter
	sync.Mutex
}

func NewMockCaseCenter() MockCaseCenter {
	return &mockCaseCenter{
		httpMock: make(map[string][]interact.HttpImposter),
		gRpcMock: make(map[string][]interact.GRpcImposter),
	}
}

// AddHttp add something http mock
func (m *mockCaseCenter) AddHttp(uniqID string, mockHttp ...interact.HttpImposter) error {
	m.Lock()
	defer m.Unlock()
	m.httpMock[uniqID] = append(m.httpMock[uniqID], mockHttp...)
	return nil
}

// AddGlobalHttp add something global http mock
func (m *mockCaseCenter) AddGlobalHttp(mockHttp ...interact.HttpImposter) {
	m.Lock()
	defer m.Unlock()
	m.httpMock[globalHttpID] = append(m.httpMock[globalHttpID], mockHttp...)
}

// GetGlobalHttp get something global http mock
func (m *mockCaseCenter) GetGlobalHttp() []interact.HttpImposter {
	m.Lock()
	defer m.Unlock()
	return m.httpMock[globalHttpID]
}

// GetHttp get something http mock
func (m *mockCaseCenter) GetHttp(uniqID string) ([]interact.HttpImposter, bool) {
	m.Lock()
	defer m.Unlock()
	ret, ok := m.httpMock[uniqID]
	return ret, ok
}

// GetAllHttp get all mock, including global mocks
func (m *mockCaseCenter) GetAllHttp() (ret []interact.HttpImposter) {
	m.Lock()
	defer m.Unlock()
	for _, mocks := range m.httpMock {
		ret = append(ret, mocks...)
	}
	return
}

// UnLoadAllHttp unload mock all of a Case
func (m *mockCaseCenter) UnLoadAllHttp() {
	m.httpMock = make(map[string][]interact.HttpImposter)
}

// UnLoadAllGlobalHttp unload mock all global of a Case
func (m *mockCaseCenter) UnLoadAllGlobalHttp() {
	m.Lock()
	defer m.Unlock()
	m.httpMock[globalHttpID] = make([]interact.HttpImposter, 0)
}

// UnLoadHttp unload mock specified
func (m *mockCaseCenter) UnLoadHttp(uniqID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.httpMock, uniqID)
}

// UnloadHttpByIdList unload the mock with Id in httpIdList
func (m *mockCaseCenter) UnloadHttpByIdList(uniqID string, httpIdList []string) {
	m.Lock()
	defer m.Unlock()
	existHttpMock := m.httpMock[uniqID]                                   // get a case mock.
	finalHttpMock := make([]interact.HttpImposter, 0, len(existHttpMock)) // save the Mock that is not a httpIdList.
	// match the case's http mock by httpIdList.
	for _, httpMock := range existHttpMock {
		unload := false
		for _, httpId := range httpIdList {
			if httpMock.Id == httpId {
				unload = true
				break
			}
		}
		// filter the contents of a httpIdList
		if !unload {
			finalHttpMock = append(finalHttpMock, httpMock)
		}
	}
	// the rest of the content is not in the httpIdList
	m.httpMock[uniqID] = finalHttpMock
}

// AddGRpc add something gRpc mock
func (m *mockCaseCenter) AddGRpc(uniqID string, mockGRpc ...interact.GRpcImposter) error {
	m.Lock()
	defer m.Unlock()
	m.gRpcMock[uniqID] = append(m.gRpcMock[uniqID], mockGRpc...)
	return nil
}

// GetGRpc get something gRpc mock
func (m *mockCaseCenter) GetGRpc(uniqID string) ([]interact.GRpcImposter, bool) {
	m.Lock()
	defer m.Unlock()
	ret, ok := m.gRpcMock[uniqID]
	return ret, ok
}

// GetAllGRpc get all gRpc mock
func (m *mockCaseCenter) GetAllGRpc() (ret []interact.GRpcImposter) {
	m.Lock()
	defer m.Unlock()
	for _, mocks := range m.gRpcMock {
		ret = append(ret, mocks...)
	}
	return
}

// UnLoadGRpc unload specified gRpc mock
func (m *mockCaseCenter) UnLoadGRpc(uniqID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.gRpcMock, uniqID)
}

// UnloadGRpcByIdList unload the mock with Id in grpcIdList
func (m *mockCaseCenter) UnloadGRpcByIdList(uniqID string, grpcIdList []string) {
	m.Lock()
	defer m.Unlock()
	existGRpcMock := m.gRpcMock[uniqID]                                   // get a case mock.
	finalGRpcMock := make([]interact.GRpcImposter, 0, len(existGRpcMock)) // save the Mock that is not a grpcIdList.
	// match the case's http mock by grpcIdList.
	for _, grpcMock := range existGRpcMock {
		unload := false
		for _, grpcId := range grpcIdList {
			if grpcMock.Id == grpcId {
				unload = true
				break
			}
		}
		// filter the contents of a grpcIdList
		if !unload {
			finalGRpcMock = append(finalGRpcMock, grpcMock)
		}
	}
	// the rest of the content is not in the grpcIdList
	m.gRpcMock[uniqID] = finalGRpcMock
}

// AddGlobalGRpc add something global grpc mock
func (m *mockCaseCenter) AddGlobalGRpc(mockGRpc ...interact.GRpcImposter) {
	m.Lock()
	defer m.Unlock()
	m.gRpcMock[globalGRpcID] = append(m.gRpcMock[globalGRpcID], mockGRpc...)
}

// GetGlobalGRpc get something global grpc mock
func (m *mockCaseCenter) GetGlobalGRpc() []interact.GRpcImposter {
	m.Lock()
	defer m.Unlock()
	return m.gRpcMock[globalGRpcID]
}
