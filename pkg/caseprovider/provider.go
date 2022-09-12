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
	"github.com/alsritter/middlebaby/pkg/types/interact"
	"github.com/alsritter/middlebaby/pkg/types/mbcase"
)

type Provider interface {
	// GetAllCaseFromItfName Get all cases form the interface serviceName.
	GetAllCaseFromItfName(serviceName string) []*mbcase.CaseTask
	GetAllCaseFromCaseName(serviceName, caseName string) *mbcase.CaseTask

	GetItfInfoFromItfName(serviceName string) *mbcase.TaskInfo
	// GetAllItfInfo Get all interface info.
	GetAllItfInfo() []*mbcase.TaskInfo
	// GetAllItf Get all interface.
	GetAllItf() []*mbcase.ItfTask

	// GetAllItfWithFileInfo  the interface that carries the file information
	GetAllItfWithFileInfo() []*mbcase.ItfTaskWithFileInfo

	// GetItfSetupCommand Get the Setup Commands of a type under the interface.
	GetItfSetupCommand(serviceName string) []*mbcase.Command
	// GetItfTearDownCommand Get the TearDown Commands of a type under the interface.
	GetItfTearDownCommand(serviceName string) []*mbcase.Command

	GetCaseSetupCommand(serviceName, caseName string) []*mbcase.Command
	GetCaseTearDownCommand(serviceName, caseName string) []*mbcase.Command

	GetMockCasesFromGlobals() []*interact.ImposterMockCase
	GetMockCasesFromItf(serviceName string) []*interact.ImposterMockCase
	GetMockCasesFromCase(serviceName, caseName string) []*interact.ImposterMockCase
}
