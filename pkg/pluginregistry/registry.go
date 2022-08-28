package pluginregistry

import (
	"sync"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

type Registry interface {
	EnvPlugins() []EnvPlugin
	RegisterEnvPlugin(...EnvPlugin) error

	AssertPlugins() []AssertPlugin
	RegisterAssertPlugin(...AssertPlugin) error
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

func New(cfg *Config, logger logger.Logger) (Registry, error) {
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
func (b *BasicRegistry) RegisterAssertPlugin(plugins ...AssertPlugin) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.assertPlugins = append(b.assertPlugins, plugins...)
	return nil
}

// RegisterEnvPlugin implements Registry
func (b *BasicRegistry) RegisterEnvPlugin(plugins ...EnvPlugin) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.envPlugins = append(b.envPlugins, plugins...)
	return nil
}
