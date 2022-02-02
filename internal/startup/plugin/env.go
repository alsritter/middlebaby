package plugin

import "github.com/alsritter/middlebaby/internal/file/config"

type Env interface {
	GetConfig() *config.Config

	GetAppArgs() []string
	GetAppPath() string
	GetTargetServeAdder() string
	GetMustRunTearDown() bool
}
