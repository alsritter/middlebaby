package cirunner

import (
	"github.com/alsritter/middlebaby/internal/startup/plugin"
	"github.com/alsritter/middlebaby/internal/task"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Plugin struct {
	taskService *task.TaskService
	log         logger.Logger
}

func NewCIRunnerPlugin(env plugin.Env, taskService *task.TaskService, log logger.Logger) *Plugin {
	return &Plugin{
		taskService: taskService,
	}
}

func (s *Plugin) Run() error {
	taskCaseMap := s.taskService.GetAllTestCase()
	mustRunTearDown := true
	for _, testCaseType := range []string{task.TestCaseTypeGRpc, task.TestCaseTypeHTTP} {
		t := taskCaseMap[testCaseType]
		if t == nil {
			continue
		}

		interfaceList := t.GetTaskCaseTree()
		for _, iFace := range interfaceList {
			for _, caseName := range iFace.CaseList {
				if err := s.taskService.Run(testCaseType, caseName, &mustRunTearDown); err != nil {
					s.log.Error(nil, "execute failure ", caseName, err.Error())
				} else {
					s.log.Debug(nil, "execute successfully ", caseName)
				}
			}
		}
	}
	return nil
}

func (s *Plugin) Name() string {
	return "ci-runner-plugin"
}
