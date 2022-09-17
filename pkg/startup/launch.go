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
	"github.com/alsritter/middlebaby/pkg/captureserver"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/messagepush"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/javascript"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/mysql"
	"github.com/alsritter/middlebaby/pkg/pluginregistry/assertprovid/redis"
	envmysql "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/mysql"
	envredis "github.com/alsritter/middlebaby/pkg/pluginregistry/envprovid/redis"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/web"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
)

func Startup(ctx *mbcontext.Context, cfg *Config, log logger.Logger, loader caseprovider.CaseLoader) error {
	pluginRegistry, err := pluginregistry.New(log, cfg.PluginRegistry)
	if err != nil {
		return err
	}

	storageProvider := storageprovider.New(log, cfg.Storage)
	pluginRegistry.RegisterEnvPlugin(
		envmysql.New(storageProvider, log),
		envredis.New(storageProvider, log))
	pluginRegistry.RegisterAssertPlugin(
		mysql.New(storageProvider, log),
		redis.New(storageProvider, log),
		javascript.New(log))

	log.Info(nil, "start loading case...")
	caseProvider, err := caseprovider.New(log, cfg.CaseProvider, loader)
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
	// mockServer := mockserver.New(log, cfg.MockServer, apiManager, protoProvider)

	msgPush := messagepush.New(log, cfg.MessagePush)

	log.Info(nil, "* start to start messagepush server")
	if err = msgPush.Start(ctx); err != nil {
		return err
	}

	captureServer := captureserver.New(log, cfg.CaptureServer, protoProvider, msgPush)
	taskServer := taskserver.New(log, cfg.TaskService, caseProvider, protoProvider, apiManager, pluginRegistry)
	targetProcess := targetprocess.New(log, cfg.TargetProcess, captureServer)

	webService := web.New(log, cfg.WebService, apiManager, caseProvider, protoProvider, taskServer, targetProcess)

	log.Info(nil, "* start to start captureServer")
	if err = captureServer.Start(ctx); err != nil {
		return err
	}

	log.Info(nil, "* start to start webService")
	if err = webService.Start(ctx); err != nil {
		return err
	}

	log.Info(nil, "* start to start targetProcess")
	if err = targetProcess.Start(ctx); err != nil {
		return err
	}

	ctx.Wait()
	return nil
}
