package startup

import (
	"fmt"
	"os"
	"os/exec"

	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/startup/plugin"
)

type TargetProcess struct {
	env plugin.Env
}

func NewTargetProcess(env plugin.Env) *TargetProcess {
	return &TargetProcess{
		env: env,
	}
}

// start the service to be tested
func (t *TargetProcess) Run() error {
	if t.env.GetAppPath() == "" {
		log.Fatal("The target application cannot be empty!")
		return nil
	}

	if _, err := os.Stat(t.env.GetAppPath()); err != nil {
		log.Fatal("target app err: ", err)
	}

	command := exec.Command(t.env.GetAppPath())

	port := t.env.GetConfig().Port

	parentEnv := os.Environ()
	// set target application proxy path.
	parentEnv = append(parentEnv, fmt.Sprintf("HTTP_PROXY=http://127.0.0.1:%d", port))
	parentEnv = append(parentEnv, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", port))
	// https to http.
	parentEnv = append(parentEnv, fmt.Sprintf("HTTPS_PROXY=http://127.0.0.1:%d", port))
	parentEnv = append(parentEnv, fmt.Sprintf("https_proxy=http://127.0.0.1:%d", port))
	command.Env = parentEnv

	// TODO: add filter support
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		if _, isExist := err.(*exec.ExitError); !isExist {
			log.Fatal("Failed to start the program to be tested, err:", err)
		}
	}

	os.Exit(0)
	return nil
}
