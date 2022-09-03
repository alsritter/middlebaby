package caseprovider

import (
	"fmt"
	"sync"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

const (
	globalCaseID = "globalCaseID"
)

type Config struct {
	TaskFileSuffix string `yaml:"taskFileSuffix"` // the default test case suffix name. example: ".case.json"

	CaseFiles  []string `yaml:"caseFiles"`
	WatchCases bool     `yaml:"watcherCases"` // whether to enable file listening

	MockFiles []string `yaml:"mockFiles"`   // mock file.
	WatchMock bool     `yaml:"watcherMock"` // whether to enable mock file listening
}

func NewConfig() *Config {
	return &Config{
		CaseFiles:      []string{},
		MockFiles:      []string{},
		WatchMock:      true,
		WatchCases:     true,
		TaskFileSuffix: ".case.json",
	}
}

func (c *Config) Validate() error {
	// no test case and no file suffix set
	if len(c.CaseFiles) == 0 || c.TaskFileSuffix == "" {
		return fmt.Errorf("no test case files were found")
	}

	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type basicProvider struct {
	cfg *Config
	logger.Logger
	// key: serviceName
	taskInterface map[string]*InterfaceTask
	mockCases     map[string][]*interact.ImposterCase

	// all test case files. (file absolute path)
	taskFiles []string
	// all test case directory. (absolute path)
	taskDirs []string

	mux sync.RWMutex
}

func New(log logger.Logger, cfg *Config) (Provider, error) {
	b := &basicProvider{
		cfg:           cfg,
		Logger:        log.NewLogger("caseProvider"),
		taskInterface: make(map[string]*InterfaceTask),
		mockCases:     make(map[string][]*interact.ImposterCase),
	}

	return b, b.init()
}

// GetAllItf implements Provider
func (b *basicProvider) GetAllItf() []*InterfaceTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	all := make([]*InterfaceTask, 0, len(b.taskInterface))
	for _, v := range b.taskInterface {
		all = append(all, v)
	}

	return all
}

// GetItfInfoFromItfName implements Provider
func (b *basicProvider) GetItfInfoFromItfName(serviceName string) *TaskInfo {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.taskInterface[serviceName].TaskInfo
}

// GetAllCaseFromCaseName implements Provider
func (b *basicProvider) GetAllCaseFromCaseName(serviceName string, caseName string) *CaseTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	cases := b.taskInterface[serviceName].Cases
	for _, v := range cases {
		if v.Name == caseName {
			return v
		}
	}

	return nil
}

// GetAllCaseFromItfName implements Provider
func (b *basicProvider) GetAllCaseFromItfName(serviceName string) []*CaseTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.taskInterface[serviceName].Cases
}

// GetAllItfInfo implements Provider
func (b *basicProvider) GetAllItfInfo() (infos []*TaskInfo) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	for _, t := range b.taskInterface {
		infos = append(infos, t.TaskInfo)
	}

	return
}

// GetItfSetupCommand implements Provider
func (b *basicProvider) GetItfSetupCommand(serviceName string, typeName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	itf := b.taskInterface[serviceName]

	for _, c := range itf.SetUp {
		if c.TypeName == typeName {
			cms = append(cms, c)
		}
	}

	return
}

// GetItfTearDownCommand implements Provider
func (b *basicProvider) GetItfTearDownCommand(serviceName string, typeName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	itf := b.taskInterface[serviceName]

	for _, c := range itf.TearDown {
		if c.TypeName == typeName {
			cms = append(cms, c)
		}
	}

	return
}

// GetMockCasesFromCase implements Provider
func (b *basicProvider) GetMockCasesFromCase(serviceName, caseName string) (ms []*interact.ImposterCase) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	itf := b.taskInterface[serviceName]
	for _, c := range itf.Cases {
		if c.Name == caseName {
			ms = append(ms, c.Mocks...)
			return
		}
	}

	b.Warn(nil, "cannot find case with name [%s] from interface [%s]", caseName, serviceName)
	return
}

// GetMockCasesFromItf implements Provider
func (b *basicProvider) GetMockCasesFromItf(serviceName string) (ms []*interact.ImposterCase) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	itf := b.taskInterface[serviceName]
	ms = append(ms, itf.Mocks...)
	return
}

func (b *basicProvider) GetMockCasesFromGlobals() []*interact.ImposterCase {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.mockCases[globalCaseID]
}
