package task

import (
	"github.com/alsritter/middlebaby/internal/file/task_file"
	"github.com/alsritter/middlebaby/internal/proxy"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
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
