package util

import (
	"context"
	"testing"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func TestStartServiceAsync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clog := logger.NewDefault("test")
	StartServiceAsync(ctx, clog, cancel, func() error {
		// Here is the initialization project
		clog.Info(nil, "TestServer Starting...")
		return nil
	}, func() error {
		// Call if cancel is closed
		clog.Info(nil, "TestServer Closed...")
		return nil
	})

	time.Sleep(time.Second * 1)

	// close.
	cancel()

	time.Sleep(time.Second * 2)
}
