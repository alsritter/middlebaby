package caseprovider

import (
	"github.com/alsritter/middlebaby/pkg/interact"
)

type Provider interface {
	// GetAllCaseFromItfName Get all cases form the interface serviceName.
	GetAllCaseFromItfName(serviceName string) []*CaseTask
	GetAllCaseFromCaseName(serviceName, caseName string) *CaseTask

	GetItfInfoFromItfName(serviceName string) *TaskInfo
	// GetAllItfInfo Get all interface info.
	GetAllItfInfo() []*TaskInfo
	// GetAllItf Get all interface.
	GetAllItf() []*ItfTask

	// GetAllItfWithFileInfo  the interface that carries the file information
	GetAllItfWithFileInfo() []*ItfTaskWithFileInfo

	// GetItfSetupCommand Get the Setup Commands of a type under the interface.
	GetItfSetupCommand(serviceName string) []*Command
	// GetItfTearDownCommand Get the TearDown Commands of a type under the interface.
	GetItfTearDownCommand(serviceName string) []*Command

	GetCaseSetupCommand(serviceName, caseName string) []*Command
	GetCaseTearDownCommand(serviceName, caseName string) []*Command

	GetMockCasesFromGlobals() []*interact.ImposterCase
	GetMockCasesFromItf(serviceName string) []*interact.ImposterCase
	GetMockCasesFromCase(serviceName, caseName string) []*interact.ImposterCase
}
