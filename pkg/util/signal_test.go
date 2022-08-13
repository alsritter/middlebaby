package util

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func TestRegisterExitHandlers(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	stop := RegisterExitHandlers(logger.NewDefault("test"), cancel)

	// do something...
	time.Sleep(2 * time.Second)
	sendInterruptSignal()

	<-stop

	clog := logger.NewDefault("test")
	clog.Info(nil, "server closed")
}

// ctrl + c
func sendInterruptSignal() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(os.Interrupt)
}
