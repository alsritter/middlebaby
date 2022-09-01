package startup

import (
	"context"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/hashicorp/go-multierror"
)

func Startup(ctx context.Context, cancelFunc context.CancelFunc, config *Config, log logger.Logger) error {
	// TODO: remove this wg, wrap it in context.
	var wg sync.WaitGroup

	// TODO: remove here...
	if config.TargetProcess.MockPort == 0 {
		config.TargetProcess.MockPort = config.MockServer.MockPort
	}

	apiManager := apimanager.New(log, config.ApiManager)
	storageRun, err := storagerunner.New(log, storageprovider.New(log, config.Storage))
	if err != nil {
		log.Fatal(nil, multierror.Prefix(err, "runner init failed").Error())
	}

	if taskService, err := taskserver.New(log, config.TaskService, apiManager, storageRun); err != nil {
		log.Fatal(nil, multierror.Prefix(err, "task service init failed").Error())
	} else {
		log.Info(nil, "* start to start taskService")
		if err := taskService.Start(); err != nil {
			return err
		}
	}

	mockServer := mockserver.New(log, config.MockServer, apiManager)
	log.Info(nil, "* start to start mockServer")
	if err := mockServer.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	target := targetprocess.New(log, config.TargetProcess)
	log.Info(nil, "* start to start target process")
	if err := target.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	wg.Wait()

	return nil
}
