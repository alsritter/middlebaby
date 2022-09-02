package startup

import (
	"context"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/mysql"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/redis"
	envmysql "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/mysql"
	envredis "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/redis"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func Startup(ctx context.Context, cancelFunc context.CancelFunc, cfg *Config, log logger.Logger) error {
	// TODO: remove this wg, wrap it in context.
	var wg sync.WaitGroup

	// TODO: remove here...
	if cfg.TargetProcess.MockPort == 0 {
		cfg.TargetProcess.MockPort = cfg.MockServer.MockPort
	}

	pluginRegistry, err := pluginregistry.New(log, cfg.PluginRegistry)
	if err != nil {
		return err
	}
	storageProvider := storageprovider.New(log, cfg.Storage)
	assertMySQL := mysql.New(storageProvider, log)
	assertRedis := redis.New(storageProvider, log)
	envMySQL := envmysql.New(storageProvider, log)
	envMySQL := envredis.New(storageProvider, log)

	caseProvider, err := caseprovider.New(log, cfg.CaseProvider)
	protoProvider, err := protomanager.New(log, cfg.ProtoManager)
	apiManager := apimanager.New(log, cfg.ApiManager, caseProvider)

	mockServer := mockserver.New(log, cfg.MockServer, apiManager, protoProvider)
	taskServer := taskserver.New(log, cfg.TaskService, caseProvider, pluginRegistry)
	targetProcess := targetprocess.New(log, cfg.TargetProcess)

	//if err != nil {
	//	log.Fatal(nil, multierror.Prefix(err, "runner init failed").Error())
	//}
	//
	//if taskService, err := taskserver.New(log, cfg.TaskService, apiManager, storageRun); err != nil {
	//	log.Fatal(nil, multierror.Prefix(err, "task service init failed").Error())
	//} else {
	//	log.Info(nil, "* start to start taskService")
	//	if err := taskService.Start(); err != nil {
	//		return err
	//	}
	//}
	//
	//mockServer := mockserver.New(log, cfg.MockServer, apiManager)
	//log.Info(nil, "* start to start mockServer")
	//if err := mockServer.Start(ctx, cancelFunc, &wg); err != nil {
	//	return err
	//}
	//
	//target := targetprocess.New(log, cfg.TargetProcess)
	//log.Info(nil, "* start to start target process")
	//if err := target.Start(ctx, cancelFunc, &wg); err != nil {
	//	return err
	//}

	wg.Wait()

	return nil
}
