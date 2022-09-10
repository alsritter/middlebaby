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
