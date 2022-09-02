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

	pluginRegistry, err := pluginregistry.New(log, cfg.PluginRegistry)
	if err != nil {
		return err
	}

	storageProvider := storageprovider.New(log, cfg.Storage)
	pluginRegistry.RegisterEnvPlugin(envmysql.New(storageProvider, log), envredis.New(storageProvider, log))
	pluginRegistry.RegisterAssertPlugin(mysql.New(storageProvider, log), redis.New(storageProvider, log))

	caseProvider, err := caseprovider.New(log, cfg.CaseProvider)
	protoProvider, err := protomanager.New(log, cfg.ProtoManager)
	apiManager := apimanager.New(log, cfg.ApiManager, caseProvider)

	mockServer := mockserver.New(log, cfg.MockServer, apiManager, protoProvider)
	taskServer := taskserver.New(log, cfg.TaskService, caseProvider, apiManager, pluginRegistry)
	targetProcess := targetprocess.New(log, cfg.TargetProcess, mockServer)

	log.Info(nil, "* start to start mockServer")
	if err = mockServer.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	log.Info(nil, "* start to start taskServer")
	if err = taskServer.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	log.Info(nil, "* start to start targetProcess")
	if err = targetProcess.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	wg.Wait()
	return nil
}
