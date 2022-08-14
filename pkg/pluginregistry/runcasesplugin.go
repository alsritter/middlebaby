package pluginregistry

import (
	"context"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type RunCasePlugin interface {
	Plugin
	Run(context.Context, logger.Logger) error
}
