package core

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"golang.org/x/sync/errgroup"
)

func Startup(appPath string, config *config.Config) {
	if appPath == "" {
		log.Fatal("The target application cannot be empty!")
	}

	if _, err := os.Stat(appPath); err != nil {
		log.Fatal("target app err: ", err)
	}

	template(appPath, config)
}

func template(appPath string, config *config.Config) {
	group := new(errgroup.Group)
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	NewMockServeBuilder(config, group, done).Build().Run()

	command := exec.Command(appPath)
	go func() {
		switch s := <-sigs; s {
		case os.Interrupt:
			log.Debug("Received: Interrupt signal.")
		case os.Kill:
			log.Debug("Received: Kill signal.")
		default:
			log.Debugf("Received other signal: %+v", s)
		}

		done <- true
		close(done)
	}()

	port := config.Port
	parentEnv := os.Environ()
	// set target application proxy path.
	parentEnv = append(parentEnv, fmt.Sprintf("HTTP_PROXY=http://127.0.0.1:%d", port))
	parentEnv = append(parentEnv, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", port))
	parentEnv = append(parentEnv, fmt.Sprintf("HTTPS_PROXY=http://127.0.0.1:%d", port))
	parentEnv = append(parentEnv, fmt.Sprintf("https_proxy=http://127.0.0.1:%d", port))

	command.Env = parentEnv
	// TODO: add filter support
	command.Stdout = os.Stdout
	command.Stderr = os.Stdout

	if err := command.Run(); err != nil {
		if _, isExist := err.(*exec.ExitError); !isExist {
			log.Fatal("Failed to start the program to be tested, err:", err)
		}
	}

	done <- true

	if err := group.Wait(); err != nil {
		log.Error("Get errors: ", err)
	}
	os.Exit(0)
}
