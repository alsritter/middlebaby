package task

import (
	"github.com/alsritter/middlebaby/internal/file/task_file"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
	"github.com/alsritter/middlebaby/pkg/proxy"
	"github.com/alsritter/middlebaby/pkg/taskserver"
)

var _ (taskserver.ITaskRunner) = (*GRpcTaskRunner)(nil)

// TODO: do something....
type GRpcTaskRunner struct {
	list []*task_file.GRpcTask
}

func newGRpcTaskRunner(list []*task_file.GRpcTask) taskserver.ITaskRunner {
	return &GRpcTaskRunner{
		list: list,
	}
}

func (g *GRpcTaskRunner) Run(caseName string, env plugin.Env, mockCenter proxy.MockCenter, runner taskserver.Runner) error {
	return nil
}

func (g *GRpcTaskRunner) GetTaskCaseTree() []*taskserver.TaskCaseTree {
	return nil
}
