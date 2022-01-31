package task

import (
	"fmt"

	"alsritter.icu/middlebaby/internal/file/task_file"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

var _ (TaskRunner) = (*HttpTaskCaseList)(nil)

type HttpTaskCaseList struct {
	list             []*task_file.HttpTask
	TestCaseNameMap  map[string]struct{}
	InterfaceNameMap map[string]struct{}
}

func (h *HttpTaskCaseList) Run(caseName string, env plugin.Env, mockCenter proxy.MockCenter, runner Runner) error {
	var (
		testCase          *HttpTaskCase
		serverInfo        *task_file.HttpTask
		interfaceOperator *task_file.InterfaceOperator
	)

out:
	for _, testCases := range h.list {
		for _, findTestCase := range testCases.TestCases {
			// 找到执行的测试用例
			if findTestCase.Name == runCfg.caseName {
				testCase = findTestCase
				serverInfo = testCases.HttpServiceInfo
				interfaceOperator = testCases.InterfaceOperator
				break out
			}
		}
	}

	if testCase == nil {
		return fmt.Errorf("没有找到测试用例")
	}

	run := HttpTestCaseRun{
		testCase:          *testCase,
		httpServiceInfo:   *serverInfo,
		runCfg:            runCfg,
		interfaceOperator: interfaceOperator,
	}

	return run.Run()
}

func (h *HttpTaskCaseList) Add(TaskRunner) error {
	return nil
}
