package taskserver

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/util"
	taskproto "github.com/alsritter/middlebaby/proto/task"
	"google.golang.org/grpc"

	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Config struct {
	CloseTearDown   bool `yaml:"closeTearDown"`
	TaskServicePort int  `yaml:"taskServicePort"`
}

func NewConfig() *Config {
	return &Config{
		CloseTearDown: false,
	}
}

func (c *Config) Validate() error {
	if c.TaskServicePort == 0 {
		return fmt.Errorf("task service port is required")
	}
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
}

type taskService struct {
	logger.Logger
	taskproto.TaskServer // implements grpc server interface task.TaskServer

	cfg            *Config
	caseProvider   caseprovider.Provider
	apiProvider    apimanager.Provider
	protoProvider  protomanager.Provider
	pluginRegistry pluginregistry.Registry
}

// New return a TaskService
func New(log logger.Logger, cfg *Config,
	caseProvider caseprovider.Provider,
	protoProvider protomanager.Provider,
	apiProvider apimanager.Provider,
	pluginRegistry pluginregistry.Registry,
) Provider {
	return &taskService{
		cfg:            cfg,
		caseProvider:   caseProvider,
		apiProvider:    apiProvider,
		pluginRegistry: pluginRegistry,
		Logger:         log.NewLogger("task"),
	}
}

// GetAllTaskCases implements task.TaskServer
func (t *taskService) GetAllTaskCases(context.Context, *taskproto.CommonRequest) (*taskproto.GetAllTaskCasesReply, error) {
	all := t.caseProvider.GetAllItf()
	return t.toGetAllTaskCasesReply(all), nil
}

// RunSingleTaskCase implements task.TaskServer
func (t *taskService) RunSingleTaskCase(ctx context.Context, req *taskproto.RunTaskRequest) (*taskproto.RunTaskReply, error) {
	t.apiProvider.LoadCaseEnv(req.ItfName, req.CaseName)
	defer t.apiProvider.ClearCaseEnv()
	if err := t.Run(ctx, req.ItfName, req.CaseName); err != nil {
		return &taskproto.RunTaskReply{
			Status:       0,
			FailedReason: err.Error(),
		}, nil
	}

	return &taskproto.RunTaskReply{
		Status:       1,
		FailedReason: "",
	}, nil
}

func (t *taskService) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.cfg.TaskServicePort))
	if err != nil {
		t.Fatal(nil, "task service failed to listen: %v", err)
		return err
	}
	server := grpc.NewServer()
	taskproto.RegisterTaskServer(server, t)

	util.StartServiceAsync(ctx, t, cancelFunc, wg, func() error {
		t.Info(nil, "Task server started, Listen port: %d", t.cfg.TaskServicePort)
		return server.Serve(listener)
	}, func() error {
		server.GracefulStop()
		return nil
	})
	return nil
}
