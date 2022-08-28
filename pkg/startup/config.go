package startup

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	Log           *logger.Config          `yaml:"log"`
	ApiManager    *apimanager.Config      `yaml:"api"`
	TargetProcess *targetprocess.Config   `yaml:"target"`
	MockServer    *mockserver.Config      `yaml:"mock"`
	Storage       *storageprovider.Config `yaml:"storage"` // mock server needs
	CaseProvider  *caseprovider.Config    `yaml:"case"`
}

func NewConfig() *Config {
	return &Config{
		Log:           logger.NewConfig(),
		ApiManager:    apimanager.NewConfig(),
		TargetProcess: targetprocess.NewConfig(),
		MockServer:    mockserver.NewConfig(),
		Storage:       storageprovider.NewConfig(),
		CaseProvider:  caseprovider.NewConfig(),
	}
}

func (c *Config) Validate() error {
	return util.ValidateConfigs(
		c.Log,
		c.Storage,
		c.CaseProvider,
		c.ApiManager,
		c.TargetProcess,
		c.MockServer,
	)
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.Storage.RegisterFlagsWithPrefix(prefix, f)
	c.ApiManager.RegisterFlagsWithPrefix(prefix, f)
	c.MockServer.RegisterFlagsWithPrefix(prefix, f)
	c.TargetProcess.RegisterFlagsWithPrefix(prefix, f)
	c.Log.RegisterFlagsWithPrefix(prefix, f)
	c.CaseProvider.RegisterFlagsWithPrefix(prefix, f)
}
