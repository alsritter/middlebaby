/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package startup

import (
	"context"
	"sync"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/mysql"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/redis"
	envmysql "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/mysql"
	envredis "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/redis"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/web"

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

	log.Info(nil, "start loading case...")
	caseProvider, err := caseprovider.New(log, cfg.CaseProvider)
	if err != nil {
		return err
	}
	log.Info(nil, "loaded case successfully")

	log.Info(nil, "start loading proto file...")
	protoProvider, err := protomanager.New(log, cfg.ProtoManager)
	if err != nil {
		return err
	}
	log.Info(nil, "loaded proto file successfully")

	apiManager := apimanager.New(log, cfg.ApiManager, caseProvider)

	mockServer := mockserver.New(log, cfg.MockServer, apiManager, protoProvider)
	taskServer := taskserver.New(log, cfg.TaskService, caseProvider, protoProvider, apiManager, pluginRegistry)
	targetProcess := targetprocess.New(log, cfg.TargetProcess, mockServer)

	webService := web.New(log, cfg.WebService, apiManager, caseProvider, protoProvider, taskServer, targetProcess)

	log.Info(nil, "* start to start mockServer")
	if err = mockServer.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	log.Info(nil, "* start to start webService")
	if err = webService.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	log.Info(nil, "* start to start targetProcess")
	if err = targetProcess.Start(ctx, cancelFunc, &wg); err != nil {
		return err
	}

	wg.Wait()
	return nil
}
