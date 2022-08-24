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
	apiManager := apimanager.New(log, config.ApiManager)
	run, err := runner.New(log, storageprovider.New(log, config.Storage))
	if err != nil {
		log.Fatal(nil, "runner init failed %v", err)
	}

	taskService, err := taskserver.New(log, config.TaskService, apiManager, run)
	if err != nil {
		log.Fatal(nil, "task service init failed %v", err)
	}
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

	// TODO: add flag
	//env := NewRunEnv(cfg, appPath, "http://127.0.0.1:9876", true)
	//
	//mockCenter := proxy.NewMockCenter()
	//trg := targetprocess.New(env, log)
	//srv := mockserver.New(env, mockCenter, log)
	//ts, err := task.NewTaskService(env, mockCenter, newRunner(env))
	//// serve := NewCaseServe(env, mockCenter, log)
	//
	//// Mock server
	//util.StartServiceAsync(ctx, log, cancel,
	//	func() error {
	//		return srv.Start()
	//	},
	//	func() error {
	//		return srv.Close()
	//	})
	//
	//// target process
	//util.StartServiceAsync(ctx, log, cancel,
	//	func() error {
	//		return trg.Run()
	//	},
	//	func() error {
	//		return trg.Close()
	//	})
	//
	//util.StartServiceAsync(ctx, log, cancel,
	//	func() error {
	//		return trg.Run()
	//	},
	//	func() error {
	//		return nil
	//	})
	//
	//// // TODO: Changes to the plugin. This is just a test.
	//// group.Go(func() error {
	//// 	defer func() {
	//// 		if err := recover(); err != nil {
	//// 			log.Fatal(nil, "panic error:", err)
	//// 		}
	//// 	}()
	//
	//// 	time.Sleep(2 * time.Second) // FIXME: remove.
	//// 	serve.Start()
	//// 	return nil
	//// })

	return nil
}
