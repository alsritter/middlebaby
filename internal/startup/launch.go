package startup

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"alsritter.icu/middlebaby/internal/event"
	"alsritter.icu/middlebaby/internal/file/config"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"golang.org/x/sync/errgroup"
)

func Startup(appPath string, config *config.Config) {
	template(appPath, config)
}

func template(appPath string, config *config.Config) {
	group := new(errgroup.Group)
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		switch s := <-sigs; s {
		case os.Interrupt:
			log.Debug("Received: Interrupt signal.")
		default:
			log.Debugf("Received other signal: %+v", s)
		}

		event.Bus.Publish(event.CLOSE)
		done <- true
		close(done)
	}()

	// TODO: add flag
	env := NewRunEnv(config, appPath, "http://127.0.0.1:9876", true)
	mockCenter := proxy.NewMockCenter()
	trg := NewTargetProcess(env)
	srv := NewMockServe(env, mockCenter)
	serve := NewCaseServe(env, mockCenter)

	group.Go(func() error {
		go func() {
			<-done
			srv.Shutdown()
		}()

		return srv.Run()
	})

	group.Go(func() error {
		return trg.Run()
	})

	// TODO: Changes to the plugin. This is just a test.
	group.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				log.Fatal("panic error:", err)
			}
		}()

		time.Sleep(2 * time.Second) // FIXME: remove.
		serve.Run()
		return nil
	})

	group.Wait()
}
