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

package cirunner

// type Plugin struct {
// 	taskService *taskserver.Provider
// 	log         logger.Logger
// }

// func NewCIRunnerPlugin(taskService *taskserver.Provider, log logger.Logger) *Plugin {
// 	return &Plugin{
// 		log:         log.NewLogger("ci-runner-plugin"),
// 		taskService: taskService,
// 	}
// }

// func (s *Plugin) Start() error {
// 	taskCaseMap := s.taskService.GetAllTestCase()
// 	for _, testCaseType := range []string{taskserver.TestCaseTypeGRpc, taskserver.TestCaseTypeHTTP} {
// 		t := taskCaseMap[testCaseType]
// 		if t == nil {
// 			continue
// 		}

// 		interfaceList := t.GetTaskCaseTree()
// 		for _, iFace := range interfaceList {
// 			for _, caseName := range iFace.CaseList {
// 				if err := s.taskService.Run(testCaseType, caseName); err != nil {
// 					s.log.Error(nil, "execute failure ", caseName, err.Error())
// 				} else {
// 					s.log.Debug(nil, "execute successfully ", caseName)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func (s *Plugin) Name() string {
// 	return "ci-runner-plugin"
// }
