package http_runner

import (
	"fmt"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

var _ runner.ITaskRunner = (*HttpTaskRunner)(nil)

// HttpTaskRunner save all HTTP task server.
// a runner contains multiple interface (taskserver == interface)
// a interface contains multiple cases
type HttpTaskRunner struct {
	list             []*task_file.HttpTask // a HttpTask contains multiple Case.
	TestCaseNameMap  map[string]struct{}
	InterfaceNameMap map[string]struct{}
	log              logger.Logger
}

func New(list []*task_file.HttpTask, log logger.Logger) runner.ITaskRunner {
	r := &HttpTaskRunner{
		list:             make([]*task_file.HttpTask, 0),
		TestCaseNameMap:  make(map[string]struct{}),
		InterfaceNameMap: make(map[string]struct{}),
		log:              log.NewLogger("HttpTaskRunner"),
	}

	if err := r.addList(list); err != nil {
		log.Error(nil, "add the task server error: %w")
	}
	return r
}

// Run executes the specified Case.
func (h *HttpTaskRunner) Run(caseName string, mockCenter apimanager.MockCaseCenter, runner runner.Runner) error {
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
		h.log.Warn(nil, "no interfaceOperator found")
		interfaceOperator = &task_file.InterfaceOperator{}
	}

	return newHttpTaskCase(*interfaceOperator, *serverInfo, *testCase, runner, mockCenter, h.log).Run()
}

func (h *HttpTaskRunner) GetTaskCaseTree() []*task_file.TaskCaseTree {
	var tree []*task_file.TaskCaseTree
	for _, service := range h.list {
		t := &task_file.TaskCaseTree{CaseList: make([]string, 0, len(service.Cases))}
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
