package startup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	"github.com/alsritter/middlebaby/internal/event"
	"github.com/alsritter/middlebaby/internal/log"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
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
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	event.Bus.Subscribe(event.CLOSE, func() {
		if err := kill(command); err != nil {
			log.Error("kill error: ", err)
		}
	})

	if err := command.Run(); err != nil {
		if _, isExist := err.(*exec.ExitError); !isExist {
			log.Fatal("Failed to start the program to be tested, err:", err)
		}
	}

	os.Exit(0)
	return nil
}

// end child process
// reference: https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
func kill(cmd *exec.Cmd) error {
	k := func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	switch runtime.GOOS {
	case "darwin":
		return k()
	case "linux":
		return k()
	case "windows":
		kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
		kill.Stderr = os.Stderr
		kill.Stdout = os.Stdout
		return kill.Run()
	}

	return nil
}
