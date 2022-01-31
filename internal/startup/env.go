package startup

import "alsritter.icu/middlebaby/internal/file/config"

type runEnv struct {
	config           *config.Config
	appPath          string
	targetServeAdder string
	mustRunTearDown  bool
}

func NewRunEnv(
	config *config.Config,
	appPath string,
	targetServeAdder string,
	mustRunTearDown bool,
) *runEnv {
	return &runEnv{config: config,
		appPath:          appPath,
		targetServeAdder: targetServeAdder,
		mustRunTearDown:  mustRunTearDown,
	}
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

func (r *runEnv) GetTargetServeAdder() string {
	return r.targetServeAdder
}

func (r *runEnv) GetMustRunTearDown() bool {
	return r.mustRunTearDown
}
