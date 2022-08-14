package startup

import (
	"context"
	"errors"
	"time"

	"github.com/alsritter/middlebaby/internal/file/config"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
	"github.com/alsritter/middlebaby/internal/task"
	"github.com/alsritter/middlebaby/pkg/proxy"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Startup(appPath string, cfg *config.Config) {
	log, err := logger.New(logger.NewConfig(), "main")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := util.RegisterExitHandlers(log, cancel)
	defer cancel()

	// TODO: add flag
	env := NewRunEnv(cfg, appPath, "http://127.0.0.1:9876", true)

	mockCenter := proxy.NewMockCenter()
	trg := NewTargetProcess(env, log)
	srv := NewMockServe(env, mockCenter, log)
	ts, err := task.NewTaskService(env, mockCenter, newRunner(env))
	// serve := NewCaseServe(env, mockCenter, log)

	// Mock server
	util.StartServiceAsync(ctx, log, cancel,
		func() error {
			return srv.Run()
		},
		func() error {
			return srv.Close()
		})

	// target process
	util.StartServiceAsync(ctx, log, cancel,
		func() error {
			return trg.Run()
		},
		func() error {
			return trg.Close()
		})

	util.StartServiceAsync(ctx, log, cancel,
		func() error {
			return trg.Run()
		},
		func() error {
			return nil
		})

	// // TODO: Changes to the plugin. This is just a test.
	// group.Go(func() error {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			log.Fatal(nil, "panic error:", err)
	// 		}
	// 	}()

	// 	time.Sleep(2 * time.Second) // FIXME: remove.
	// 	serve.Run()
	// 	return nil
	// })

	<-stop
	log.Info(nil, "Goodbye")
}
