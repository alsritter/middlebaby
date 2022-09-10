/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	taskInterface map[string]*ItfTask
	mockCases     map[string][]*interact.ImposterCase

	taskWithFileInfo map[string]*ItfTaskWithFileInfo // serviceName relative file info.

	// all test case files. (file absolute path)
	taskFiles []string
	// all test case directory. (absolute path)
	taskDirs []string

	mux sync.RWMutex
}

func New(log logger.Logger, cfg *Config) (Provider, error) {
	b := &basicProvider{
		cfg:              cfg,
		Logger:           log.NewLogger("case"),
		taskInterface:    make(map[string]*ItfTask),
		taskWithFileInfo: make(map[string]*ItfTaskWithFileInfo),
		mockCases:        make(map[string][]*interact.ImposterCase),
	}

	return b, b.init()
}

// GetAllItfWithFileInfo implements Provider
func (b *basicProvider) GetAllItfWithFileInfo() []*ItfTaskWithFileInfo {
	b.mux.RLock()
	defer b.mux.RUnlock()

	all := make([]*ItfTaskWithFileInfo, 0, len(b.taskWithFileInfo))
	for _, v := range b.taskWithFileInfo {
		all = append(all, v)
	}
	return all
}

// GetAllItf implements Provider
func (b *basicProvider) GetAllItf() []*ItfTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	all := make([]*ItfTask, 0, len(b.taskInterface))
	for _, v := range b.taskInterface {
		all = append(all, v)
	}

	return all
}

// GetItfInfoFromItfName implements Provider
func (b *basicProvider) GetItfInfoFromItfName(serviceName string) *TaskInfo {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if ti, ok := b.taskInterface[serviceName]; ok {
		return ti.TaskInfo
	}
	return nil
}

// GetAllCaseFromCaseName implements Provider
func (b *basicProvider) GetAllCaseFromCaseName(serviceName, caseName string) *CaseTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if ti, ok := b.taskInterface[serviceName]; ok {
		for _, v := range ti.Cases {
			if v.Name == caseName {
				return v
			}
		}
	}
	return nil
}

// GetAllCaseFromItfName implements Provider
func (b *basicProvider) GetAllCaseFromItfName(serviceName string) []*CaseTask {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if ti, ok := b.taskInterface[serviceName]; ok {
		return ti.Cases
	}
	return nil
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
func (b *basicProvider) GetItfSetupCommand(serviceName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		cms = append(cms, itf.SetUp...)
	}
	return
}

// GetItfTearDownCommand implements Provider
func (b *basicProvider) GetItfTearDownCommand(serviceName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		cms = append(cms, itf.TearDown...)
	}
	return
}

// GetCaseSetupCommand implements Provider
func (b *basicProvider) GetCaseSetupCommand(serviceName, caseName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		for _, c := range itf.Cases {
			if c.Name == caseName {
				cms = append(cms, c.SetUp...)
				return
			}
		}
	}
	return
}

// GetCaseTearDownCommand implements Provider
func (b *basicProvider) GetCaseTearDownCommand(serviceName, caseName string) (cms []*Command) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		for _, c := range itf.Cases {
			if c.Name == caseName {
				cms = append(cms, c.TearDown...)
				return
			}
		}
	}
	return
}

// GetMockCasesFromCase implements Provider
func (b *basicProvider) GetMockCasesFromCase(serviceName, caseName string) (ms []*interact.ImposterCase) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		for _, c := range itf.Cases {
			if c.Name == caseName {
				ms = append(ms, c.Mocks...)
				return
			}
		}
	}

	b.Warn(nil, "cannot find case with name [%s] from interface [%s]", caseName, serviceName)
	return
}

// GetMockCasesFromItf implements Provider
func (b *basicProvider) GetMockCasesFromItf(serviceName string) (ms []*interact.ImposterCase) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if itf, ok := b.taskInterface[serviceName]; ok {
		ms = append(ms, itf.Mocks...)
	}
	return
}

func (b *basicProvider) GetMockCasesFromGlobals() []*interact.ImposterCase {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.mockCases[globalCaseID]
}
