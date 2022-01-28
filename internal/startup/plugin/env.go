package plugin

import "alsritter.icu/middlebaby/internal/config"

type Env interface {
	GetConfig() *config.Config
	GetAppArgs() []string
	GetAppPath() string
}
