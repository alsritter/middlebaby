package caseprovider

import (
	"sync"

	"github.com/alsritter/middlebaby/pkg/interact"
)

// MockCaseCenter save all type mock.
type MockCaseProvider interface {
	// AddMockCase add something mock case
	AddMockCase(uniqID string, mockCase ...interact.ImposterCase) error
	// AddGlobalMockCase add something global mock case
	AddGlobalMockCase(mockCase ...interact.ImposterCase)
	// GetGlobalMockCase get something global mock case
	GetGlobalMockCase() []interact.ImposterCase
	// GetMockCase get something Case mock
	GetMockCase(uniqID string) ([]interact.ImposterCase, bool)
	// GetAllMockCase get all mock, including global mocks
	GetAllMockCase() []interact.ImposterCase
	// UnLoadAllMockCase unload mock all of a Case
	UnLoadAllMockCase()
	// UnLoadAllGlobalMockCase unload mock all global of a Case
	UnLoadAllGlobalMockCase()
	// UnLoadMockCase unload mock specified
	UnLoadMockCase(uniqID string)
	// UnloadMockCaseByIdList unload the mock with Id in caseIdList
	UnloadMockCaseByIdList(uniqID string, caseIdList []string)
}

// mockCaseCenter storage all mock cases
type mockCaseCenter struct {
	cases map[string][]interact.ImposterCase
	sync.Mutex
}

func NewMockCaseCenter() MockCaseProvider {
	return &mockCaseCenter{
		cases: make(map[string][]interact.ImposterCase),
	}
}

// AddMockCase add something Case mock
func (m *mockCaseCenter) AddMockCase(uniqID string, mockCase ...interact.ImposterCase) error {
	m.Lock()
	defer m.Unlock()
	m.cases[uniqID] = append(m.cases[uniqID], mockCase...)
	return nil
}

// AddGlobalMockCase add something global Case mock
func (m *mockCaseCenter) AddGlobalMockCase(mockCase ...interact.ImposterCase) {
	m.Lock()
	defer m.Unlock()
	m.cases[globalCaseID] = append(m.cases[globalCaseID], mockCase...)
}

// GetGlobalMockCase get something global Case mock
func (m *mockCaseCenter) GetGlobalMockCase() []interact.ImposterCase {
	m.Lock()
	defer m.Unlock()
	return m.cases[globalCaseID]
}

// GetMockCase get something Case mock
func (m *mockCaseCenter) GetMockCase(uniqID string) ([]interact.ImposterCase, bool) {
	m.Lock()
	defer m.Unlock()
	ret, ok := m.cases[uniqID]
	return ret, ok
}

// GetAllMockCase get all mock, including global mocks
func (m *mockCaseCenter) GetAllMockCase() (ret []interact.ImposterCase) {
	m.Lock()
	defer m.Unlock()
	for _, mocks := range m.cases {
		ret = append(ret, mocks...)
	}
	return
}

// UnLoadAllMockCase unload mock all of a Case
func (m *mockCaseCenter) UnLoadAllMockCase() {
	m.cases = make(map[string][]interact.ImposterCase)
}

// UnLoadAllGlobalMockCase unload mock all global of a Case
func (m *mockCaseCenter) UnLoadAllGlobalMockCase() {
	m.Lock()
	defer m.Unlock()
	m.cases[globalCaseID] = make([]interact.ImposterCase, 0)
}

// UnLoadMockCase unload mock specified
func (m *mockCaseCenter) UnLoadMockCase(uniqID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.cases, uniqID)
}

// UnloadMockCaseByIdList unload the mock with Id in CaseIdList
func (m *mockCaseCenter) UnloadMockCaseByIdList(uniqID string, CaseIdList []string) {
	m.Lock()
	defer m.Unlock()
	existcases := m.cases[uniqID]                                   // get a case mock.
	finalcases := make([]interact.ImposterCase, 0, len(existcases)) // save the Mock that is not a CaseIdList.
	// match the case's Case mock by CaseIdList.
	for _, cases := range existcases {
		unload := false
		for _, CaseId := range CaseIdList {
			if cases.Id == CaseId {
				unload = true
				break
			}
		}
		// filter the contents of a CaseIdList
		if !unload {
			finalcases = append(finalcases, cases)
		}
	}
	// the rest of the content is not in the CaseIdList
	m.cases[uniqID] = finalcases
}
