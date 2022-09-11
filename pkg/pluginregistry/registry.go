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

package pluginregistry

import (
	"sync"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

type Registry interface {
	EnvPlugins() []EnvPlugin
	RegisterEnvPlugin(...EnvPlugin)

	AssertPlugins() []AssertPlugin
	RegisterAssertPlugin(...AssertPlugin)
}

// Config defines the config structure
type Config struct{}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return nil
}

type BasicRegistry struct {
	cfg *Config

	envPlugins    []EnvPlugin
	assertPlugins []AssertPlugin
	lock          sync.Mutex

	logger.Logger
}

func New(logger logger.Logger, cfg *Config) (Registry, error) {
	service := &BasicRegistry{
		envPlugins:    []EnvPlugin{},
		assertPlugins: []AssertPlugin{},
		cfg:           cfg,
		Logger:        logger.NewLogger("pluginsRegistry"),
	}
	return service, nil
}

// AssertPlugins implements Registry
func (b *BasicRegistry) AssertPlugins() []AssertPlugin {
	return b.assertPlugins
}

// EnvPlugins implements Registry
func (b *BasicRegistry) EnvPlugins() []EnvPlugin {
	return b.envPlugins
}

// RegisterAssertPlugin implements Registry
func (b *BasicRegistry) RegisterAssertPlugin(plugins ...AssertPlugin) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.assertPlugins = append(b.assertPlugins, plugins...)
	return
}

// RegisterEnvPlugin implements Registry
func (b *BasicRegistry) RegisterEnvPlugin(plugins ...EnvPlugin) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.envPlugins = append(b.envPlugins, plugins...)
	return
}
