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
