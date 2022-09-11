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

package taskserver

import (
	"context"
	"fmt"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/protomanager"

	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type RunTaskReply struct {
	Status       int32  `yaml:"status" json:"status"`
	FailedReason string `yaml:"failedReason" json:"failedReason"`
}

type Config struct {
	CloseTearDown    bool   `yaml:"closeTearDown"`
	TargetServeAdder string `yaml:"targetServeAdder"`
}

func NewConfig() *Config {
	return &Config{
		CloseTearDown: false,
	}
}

func (c *Config) Validate() error {
	if c.TargetServeAdder == "" {
		return fmt.Errorf("target Serve Adder cannot be empty")
	}

	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	RunSingleTaskCase(ctx context.Context, itfName, caseName string) (RunTaskReply, error)
}

type taskService struct {
	logger.Logger

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
		protoProvider:  protoProvider,
		apiProvider:    apiProvider,
		pluginRegistry: pluginRegistry,
		Logger:         log.NewLogger("task"),
	}
}

// RunSingleTaskCase implements task.TaskServer
func (t *taskService) RunSingleTaskCase(ctx context.Context, itfName, caseName string) (RunTaskReply, error) {
	if err := t.Run(ctx, itfName, caseName); err != nil {
		t.Error(map[string]interface{}{
			"InterfaceName": itfName,
			"CaseName":      caseName,
		}, err.Error())

		return RunTaskReply{
			Status:       0,
			FailedReason: err.Error(),
		}, nil
	}

	t.Info(map[string]interface{}{
		"InterfaceName": itfName,
		"CaseName":      caseName,
	}, "case assert successful")
	return RunTaskReply{
		Status:       1,
		FailedReason: "",
	}, nil
}
