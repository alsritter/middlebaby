package startup

import (
	"os"
	"os/signal"
	"syscall"

	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
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

		done <- true
		close(done)
	}()

	env := NewRunEnv(config, appPath)

	trg := NewTargetProcess(env)
	srv := NewMockServe(env)

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

	group.Wait()
}
