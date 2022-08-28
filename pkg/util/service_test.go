package util

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func TestStartServiceAsync(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clog := logger.NewDefault("test")
	var wg sync.WaitGroup
	StartServiceAsync(ctx, clog, cancel, &wg, func() error {
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

// Reference:
// https://stackoverflow.com/questions/66833138/wait-for-context-done-channel-for-cancellation-while-working-on-long-run-operati
func TestContext(t *testing.T) {
	var (
		workTimeCost  = 2 * time.Second
		cancelTimeout = 1 * time.Second
	)

	ctx, cancel := context.WithCancel(context.Background())

	var (
		data   int
		readCh = make(chan struct{})
	)

	go func() {
		defer close(readCh)
		t.Log("blocked to read data")
		// fake long i/o operations
		time.Sleep(workTimeCost)
		data = 10
		t.Log("done read data")
	}()

	// fake cancel is called from the other routine (it's actually not caused by timeout)
	time.AfterFunc(cancelTimeout, cancel)

	select {
	case <-ctx.Done():
		t.Log("cancelled")
		return
	case <-readCh:
		break
	}

	t.Log("got final data", data)
}
