package task

import (
	"alsritter.icu/middlebaby/internal/file/task_file"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

var _ (ITaskRunner) = (*GRpcTaskRunner)(nil)

// TODO: do something....
type GRpcTaskRunner struct {
	list []*task_file.GRpcTask
}

func newGRpcTaskRunner(list []*task_file.GRpcTask) ITaskRunner {
	return &GRpcTaskRunner{
		list: list,
	}
}

func (g *GRpcTaskRunner) Run(caseName string, env plugin.Env, mockCenter proxy.MockCenter, runner Runner) error {
	return nil
}

func (g *GRpcTaskRunner) GetTaskCaseTree() []*TaskCaseTree {
	return nil
}
