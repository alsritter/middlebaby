package startup

import "alsritter.icu/middlebaby/internal/config"

type runEnv struct {
	config  *config.Config
	appPath string
}

func NewRunEnv(config *config.Config, appPath string) *runEnv {
	return &runEnv{config: config, appPath: appPath}
}

func (r *runEnv) GetConfig() *config.Config {
	return r.config
}

func (r *runEnv) GetAppPath() string {
	return r.appPath
}

func (r *runEnv) GetAppArgs() []string {
	return []string{}
}
