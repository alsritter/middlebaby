package startup

import (
	"context"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func Startup(ctx context.Context, cancelFunc context.CancelFunc, config *Config, log logger.Logger) error {
	// TODO: remove here...
	if config.TargetProcess.MockPort == 0 {
		config.TargetProcess.MockPort = config.MockServer.MockPort
	}

	apiManager := apimanager.New(log, config.ApiManager)
	run, err := runner.New(log, storageprovider.New(log, config.Storage))
	if err != nil {
		log.Fatal(nil, "runner init failed %v", err)
	}

	taskService, err := taskserver.New(log, config.TaskService, apiManager, run)
	if err != nil {
		log.Fatal(nil, "task service init failed %v", err)
	}
	log.Info(nil, "* start to start taskService")
	if err := taskService.Start(); err != nil {
		return err
	}

	mockServer := mockserver.New(log, config.MockServer, apiManager)
	log.Info(nil, "* start to start mockServer")
	if err := mockServer.Start(ctx, cancelFunc); err != nil {
		return err
	}

	target := targetprocess.New(log, config.TargetProcess)
	log.Info(nil, "* start to start target process")
	if err := target.Start(ctx, cancelFunc); err != nil {
		return err
	}
	return nil
}
