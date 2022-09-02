package startup

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/pluginregistry"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/storageprovider"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	Log            *logger.Config          `yaml:"log"`
	ApiManager     *apimanager.Config      `yaml:"api"`
	TargetProcess  *targetprocess.Config   `yaml:"target"`
	MockServer     *mockserver.Config      `yaml:"mock"`
	TaskService    *taskserver.Config      `yaml:"task"`
	Storage        *storageprovider.Config `yaml:"storage"`
	CaseProvider   *caseprovider.Config    `yaml:"case"`
	ProtoManager   *protomanager.Config    `yaml:"proto"`
	PluginRegistry *pluginregistry.Config  `yaml:"plugin"`
}

func NewConfig() *Config {
	return &Config{
		Log:            logger.NewConfig(),
		ApiManager:     apimanager.NewConfig(),
		TargetProcess:  targetprocess.NewConfig(),
		MockServer:     mockserver.NewConfig(),
		Storage:        storageprovider.NewConfig(),
		CaseProvider:   caseprovider.NewConfig(),
		ProtoManager:   protomanager.NewConfig(),
		TaskService:    taskserver.NewConfig(),
		PluginRegistry: pluginregistry.NewConfig(),
	}
}

func (c *Config) Validate() error {
	return util.ValidateConfigs(
		c.Log,
		c.Storage,
		c.ApiManager,
		c.MockServer,
		c.TaskService,
		c.ProtoManager,
		c.CaseProvider,
		c.TargetProcess,
		c.PluginRegistry,
	)
}

func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.Log.RegisterFlagsWithPrefix(prefix, f)
	c.Storage.RegisterFlagsWithPrefix(prefix, f)
	c.ApiManager.RegisterFlagsWithPrefix(prefix, f)
	c.MockServer.RegisterFlagsWithPrefix(prefix, f)
	c.TaskService.RegisterFlagsWithPrefix(prefix, f)
	c.CaseProvider.RegisterFlagsWithPrefix(prefix, f)
	c.ProtoManager.RegisterFlagsWithPrefix(prefix, f)
	c.TargetProcess.RegisterFlagsWithPrefix(prefix, f)
	c.PluginRegistry.RegisterFlagsWithPrefix(prefix, f)
}
