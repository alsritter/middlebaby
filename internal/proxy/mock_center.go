package proxy

import (
	"sync"

	"alsritter.icu/middlebaby/internal/file/common"
)

// save all type mock.
type MockCenter interface {
	// add something http mock
	AddHttp(uniqID string, mockHttp ...common.HttpImposter) error
	// add something global http mock
	AddGlobalHttp(mockHttp ...common.HttpImposter)
	// get something global http mock
	GetGlobalHttp() []common.HttpImposter
	// get something http mock
	GetHttp(uniqID string) ([]common.HttpImposter, bool)
	// get all mock, including global mocks
	GetAllHttp() []common.HttpImposter
	// unload mock all of a Case
	UnLoadAllHttp()
	// unload mock specified
	UnLoadHttp(uniqID string)
	// unload the mock with Id in httpIdList
	UnloadHttpByIdList(uniqID string, httpIdList []string)

	// add something gRpc mock
	AddGRpc(uniqID string, mockGRpc ...common.GRpcImposter) error
	// get something gRpc mock
	GetGRpc(uniqID string) ([]common.GRpcImposter, bool)
	// get all gRpc mock
	GetAllGRpc() []common.GRpcImposter
	// unload specified gRpc mock
	UnLoadGRpc(uniqID string)
	// unload the mock with Id in grpcIdList
	UnloadGRpcByIdList(uniqID string, grpcIdList []string)
	// add something global grpc mock
	AddGlobalGRpc(mockGRpc ...common.GRpcImposter)
	// get something global grpc mock
	GetGlobalGRpc() []common.GRpcImposter
}

const (
	globalHttpID = "globalHttpID"
	globalGRpcID = "globalGRpcID"
)

// mockCenter handler all mock
type mockCenter struct {
	httpMock map[string][]common.HttpImposter
	gRpcMock map[string][]common.GRpcImposter
	sync.Mutex
}

func NewMockCenter() MockCenter {
	return &mockCenter{
		httpMock: make(map[string][]common.HttpImposter),
		gRpcMock: make(map[string][]common.GRpcImposter),
	}
}

// add something http mock
func (m *mockCenter) AddHttp(uniqID string, mockHttp ...common.HttpImposter) error {
	m.Lock()
	defer m.Unlock()
	m.httpMock[uniqID] = append(m.httpMock[uniqID], mockHttp...)
	return nil
}

// add something global http mock
func (m *mockCenter) AddGlobalHttp(mockHttp ...common.HttpImposter) {
	m.Lock()
	defer m.Unlock()
	m.httpMock[globalHttpID] = append(m.httpMock[globalHttpID], mockHttp...)
}

// get something global http mock
func (m *mockCenter) GetGlobalHttp() []common.HttpImposter {
	m.Lock()
	defer m.Unlock()
	return m.httpMock[globalHttpID]
}

// get something http mock
func (m *mockCenter) GetHttp(uniqID string) ([]common.HttpImposter, bool) {
	m.Lock()
	defer m.Unlock()
	ret, ok := m.httpMock[uniqID]
	return ret, ok
}

// get all mock, including global mocks
func (m *mockCenter) GetAllHttp() (ret []common.HttpImposter) {
	m.Lock()
	defer m.Unlock()
	for _, mocks := range m.httpMock {
		ret = append(ret, mocks...)
	}
	return
}

// unload mock all of a Case
func (m *mockCenter) UnLoadAllHttp() {
	m.httpMock = make(map[string][]common.HttpImposter)
}

// unload mock specified
func (m *mockCenter) UnLoadHttp(uniqID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.httpMock, uniqID)
}

// unload the mock with Id in httpIdList
func (m *mockCenter) UnloadHttpByIdList(uniqID string, httpIdList []string) {
	m.Lock()
	defer m.Unlock()
	existHttpMock := m.httpMock[uniqID]                                 // get a case mock.
	finalHttpMock := make([]common.HttpImposter, 0, len(existHttpMock)) // save the Mock that is not a httpIdList.
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

// add something gRpc mock
func (m *mockCenter) AddGRpc(uniqID string, mockGRpc ...common.GRpcImposter) error {
	m.Lock()
	defer m.Unlock()
	m.gRpcMock[uniqID] = append(m.gRpcMock[uniqID], mockGRpc...)
	return nil
}

// get something gRpc mock
func (m *mockCenter) GetGRpc(uniqID string) ([]common.GRpcImposter, bool) {
	m.Lock()
	defer m.Unlock()
	ret, ok := m.gRpcMock[uniqID]
	return ret, ok
}

// get all gRpc mock
func (m *mockCenter) GetAllGRpc() (ret []common.GRpcImposter) {
	m.Lock()
	defer m.Unlock()
	for _, mocks := range m.gRpcMock {
		ret = append(ret, mocks...)
	}
	return
}

// unload specified gRpc mock
func (m *mockCenter) UnLoadGRpc(uniqID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.gRpcMock, uniqID)
}

// unload the mock with Id in grpcIdList
func (m *mockCenter) UnloadGRpcByIdList(uniqID string, grpcIdList []string) {
	m.Lock()
	defer m.Unlock()
	existGRpcMock := m.gRpcMock[uniqID]                                 // get a case mock.
	finalGRpcMock := make([]common.GRpcImposter, 0, len(existGRpcMock)) // save the Mock that is not a grpcIdList.
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

// add something global grpc mock
func (m *mockCenter) AddGlobalGRpc(mockGRpc ...common.GRpcImposter) {
	m.Lock()
	defer m.Unlock()
	m.gRpcMock[globalGRpcID] = append(m.gRpcMock[globalGRpcID], mockGRpc...)
}

// get something global grpc mock
func (m *mockCenter) GetGlobalGRpc() []common.GRpcImposter {
	m.Lock()
	defer m.Unlock()
	return m.gRpcMock[globalGRpcID]
}
