package startup

import (
	"context"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func Startup() {
	log, err := logger.New(logger.NewConfig(), "main")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := util.RegisterExitHandlers(log, cancel)
	defer cancel()

	log.Info(nil, "", ctx.Value("占位符"))

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

	<-stop
	log.Info(nil, "Goodbye")
}

//func newRunner(env plugin.Env) Runner {
//	db, err := getMysqlCon(env)
//	if err != nil {
//		log.Error("Failed to connect to the MySQL database:", err.Error())
//	}
//
//	redisPool, err := getRedisCon(env)
//	if err != nil {
//		log.Error("Failed to connect to the Redis:", err.Error())
//	}
//	runner, err := NewRunner(storage_runner2.NewMysqlRunner(db), storage_runner2.NewRedisRunner(redisPool))
//	if err != nil {
//		log.Fatal("Failed to initialize the running environment:", err)
//	}
//	return runner
//}
