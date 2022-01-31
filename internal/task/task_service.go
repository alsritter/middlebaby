package task

import (
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

type TestCaseType = string

const (
	TestCaseTypeHTTP TestCaseType = "http"
	TestCaseTypeGRpc TestCaseType = "grpc"
)

type TaskService struct {
	// the default test case suffix name.
	defaultCaseDescFileSuffix string
	// all test case description files.
	taskDescFiles []string
	// collection of all test case description files.
	// taskCases map[TestCaseType]ITestCaseRunner
	// provides an interface for use case execution
	runner Runner
	// mock center
	mockCenter proxy.MockCenter
	// configuration information required by the service.
	env plugin.Env
}
