package targetprocess

import (
	"context"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/mockserver"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Config struct {
	AppPath  string `yaml:"appPath"`
	MockPort int    `yaml:"mockPort"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.AppPath == "" {
		return fmt.Errorf("the target application cannot be empty")
	}

	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.AppPath, prefix+"target.path", c.AppPath, "target application address")
}

// Provider defines the target process interface
type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
}

type TargetProcess struct {
	cfg     *Config
	command *exec.Cmd
	log     logger.Logger
}

func New(log logger.Logger, cfg *Config, mock mockserver.Provider) Provider {
	cfg.MockPort = mock.GetPort()
	return &TargetProcess{
		cfg: cfg,
		log: log.NewLogger("target"),
	}
}

// Start the service to be tested
func (t *TargetProcess) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	util.StartServiceAsync(ctx, t.log, cancelFunc, wg, func() error {
		if _, err := os.Stat(t.cfg.AppPath); err != nil {
			return fmt.Errorf("target app err: %v", err)
		}

		t.command = exec.Command(t.cfg.AppPath)

		port := t.cfg.MockPort

		parentEnv := os.Environ()
		// set target application proxy path.
		parentEnv = append(parentEnv, fmt.Sprintf("HTTP_PROXY=http://127.0.0.1:%d", port))
		parentEnv = append(parentEnv, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", port))

		// https to http.
		parentEnv = append(parentEnv, fmt.Sprintf("HTTPS_PROXY=http://127.0.0.1:%d", port))
		parentEnv = append(parentEnv, fmt.Sprintf("https_proxy=http://127.0.0.1:%d", port))
		t.command.Env = parentEnv

		// TODO: add filter support
		t.command.Stdout = os.Stdout
		t.command.Stderr = os.Stderr
		t.command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		if err := t.command.Run(); err != nil {
			if _, isExist := err.(*exec.ExitError); !isExist {
				return fmt.Errorf("failed to start the program to be tested, err: %v", err)
			}
		}
		return nil
	}, func() error {
		if err := kill(t.command); err != nil {
			return fmt.Errorf("kill error: %v", err)
		}
		return nil
	})
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
