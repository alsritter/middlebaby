package task

import (
	"fmt"

	"alsritter.icu/middlebaby/internal/file/task_file"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

var _ (ITaskRunner) = (*HttpTaskRunner)(nil)

// save all HTTP task.
// a runner contains multiple interface (task == interface)
// a interface contains multiple cases
type HttpTaskRunner struct {
	list             []*task_file.HttpTask // a HttpTask contains multiple Case.
	TestCaseNameMap  map[string]struct{}
	InterfaceNameMap map[string]struct{}
}

func newHttpTaskRunner(list []*task_file.HttpTask) ITaskRunner {
	r := &HttpTaskRunner{
		list:             make([]*task_file.HttpTask, 0),
		TestCaseNameMap:  make(map[string]struct{}),
		InterfaceNameMap: make(map[string]struct{}),
	}

	if err := r.addList(list); err != nil {
		log.Error("add the task error: %w")
	}
	return r
}

// executes the specified Case.
func (h *HttpTaskRunner) Run(caseName string, env plugin.Env, mockCenter proxy.MockCenter, runner Runner) error {
	var (
		testCase          *task_file.HttpTaskCase
		serverInfo        *task_file.HttpTaskInfo
		interfaceOperator *task_file.InterfaceOperator
	)

	// break label, break out of the two-layer loop
out:
	for _, httpTask := range h.list {
		for _, findTestCase := range httpTask.Cases {
			// find the case.
			if findTestCase.Name == caseName {
				testCase = findTestCase
				serverInfo = httpTask.HttpTaskInfo
				interfaceOperator = httpTask.InterfaceOperator
				break out
			}
		}
	}

	if testCase == nil {
		return fmt.Errorf("no test case found")
	}

	if serverInfo == nil {
		return fmt.Errorf("no serverInfo found")
	}

	if interfaceOperator == nil {
		log.Warn("no interfaceOperator found")
		interfaceOperator = &task_file.InterfaceOperator{}
	}

	runCase := NewHttpTaskCase(*testCase, *serverInfo, runner, mockCenter, env, *interfaceOperator)
	return runCase.Run()
}

func (h *HttpTaskRunner) GetTaskCaseTree() []*TaskCaseTree {
	var tree []*TaskCaseTree
	for _, service := range h.list {
		t := &TaskCaseTree{CaseList: make([]string, 0, len(service.Cases))}
		t.InterfaceName = service.HttpTaskInfo.ServiceName
		for _, testCase := range service.Cases {
			t.CaseList = append(t.CaseList, testCase.Name)
		}

		tree = append(tree, t)
	}
	return tree
}

func (h *HttpTaskRunner) addList(list []*task_file.HttpTask) error {
	for _, httpTask := range list {
		if err := h.addToInterfaceNameMap(httpTask.ServiceName); err != nil {
			return err
		}

		if err := h.addToTestCaseNameMap(httpTask.Cases); err != nil {
			return err
		}
	}

	h.list = append(h.list, list...)
	return nil
}

func (h *HttpTaskRunner) addToInterfaceNameMap(interfaceName string) error {
	if _, ok := h.InterfaceNameMap[interfaceName]; ok {
		return fmt.Errorf("HTTP interface name duplicate: %s", interfaceName)
	} else {
		h.InterfaceNameMap[interfaceName] = struct{}{}
	}
	return nil
}

func (h *HttpTaskRunner) addToTestCaseNameMap(testCases []*task_file.HttpTaskCase) error {
	for _, testCase := range testCases {
		if _, ok := h.TestCaseNameMap[testCase.Name]; ok {
			return fmt.Errorf("HTTP case name duplicate: %s", testCase.Name)
		} else {
			h.TestCaseNameMap[testCase.Name] = struct{}{}
		}
	}
	return nil
}
