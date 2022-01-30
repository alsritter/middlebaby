package plugin

import "alsritter.icu/middlebaby/internal/file/config"

type Env interface {
	GetConfig() *config.Config
	GetAppArgs() []string
	GetAppPath() string
}
