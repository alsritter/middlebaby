package util

import (
	"context"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

// StartServiceAsync is used to start service async
func StartServiceAsync(ctx context.Context, log logger.Logger, cancelFunc context.CancelFunc, serveFn func() error, stopFn func() error) {
	if serveFn == nil {
		return
	}
	go func() {
		log.Info(nil, "starting service")
		go func() {
			if err := serveFn(); err != nil {
				log.Error(nil, "error serving service: %s", err)
			}
			if cancelFunc != nil {
				cancelFunc()
			}
		}()

		<-ctx.Done()
		log.Info(nil, "stopping service")
		if stopFn() != nil {
			log.Info(nil, "stopping service gracefully")
			if err := stopFn(); err != nil {
				log.Warn(nil, "error occurred while stopping service: %s", err)
			}
		}
		log.Info(nil, "exiting service")
	}()
}
