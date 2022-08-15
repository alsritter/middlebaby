package grpc_runner

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

var _ taskserver.ITaskRunner = (*GRpcTaskRunner)(nil)

// GRpcTaskRunner TODO: do something....
type GRpcTaskRunner struct {
	list []*task_file.GRpcTask
	log  logger.Logger
}

func New(list []*task_file.GRpcTask, log logger.Logger) taskserver.ITaskRunner {
	return &GRpcTaskRunner{
		list: list,
		log:  log,
	}
}

func (g *GRpcTaskRunner) Run(caseName string, mockCenter apimanager.ApiMockCenter, runner runner.Runner) error {
	return nil
}

func (g *GRpcTaskRunner) GetTaskCaseTree() []*taskserver.TaskCaseTree {
	return nil
}
