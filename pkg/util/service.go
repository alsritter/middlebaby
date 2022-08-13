package util

import (
	"context"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

// StartServiceAsync is used to start service async
func StartServiceAsync(ctx context.Context, logger logger.Logger, cancelFunc context.CancelFunc, serveFn func() error, stopFn func() error) {
	if serveFn == nil {
		return
	}
	go func() {
		logger.Info(nil, "starting service")
		go func() {
			if err := serveFn(); err != nil {
				logger.Error(nil, "error serving service: %s", err)
			}
			if cancelFunc != nil {
				cancelFunc()
			}
		}()
		<-ctx.Done()
		logger.Info(nil, "stopping service")
		if stopFn() != nil {
			logger.Info(nil, "stopping service gracefully")
			if err := stopFn(); err != nil {
				logger.Warn(nil, "error occurred while stopping service: %s", err)
			}
		}
		logger.Info(nil, "exiting service")
	}()
}
