package targetprocess

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Config struct {
	AppPath string `yaml:"appPath"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.AppPath, prefix+"target.path", c.AppPath, "target application address")
}

type TargetProcess struct {
	cfg     *Config
	command *exec.Cmd
	log     logger.Logger
}

func New(cfg *Config, log logger.Logger) *TargetProcess {
	return &TargetProcess{
		cfg: cfg,
		log: log,
	}
}

// Run start the service to be tested
func (t *TargetProcess) Run() error {
	if t.cfg.AppPath == "" {
		return fmt.Errorf("The target application cannot be empty!")
	}

	if _, err := os.Stat(t.cfg.AppPath); err != nil {
		return fmt.Errorf("target app err: ", err)
	}

	t.command = exec.Command(t.cfg.AppPath)

	port := "8888"

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
			return fmt.Errorf("Failed to start the program to be tested, err:", err)
		}
	}
	return nil
}

func (t *TargetProcess) Close() error {
	if err := kill(t.command); err != nil {
		return fmt.Errorf("kill error: ", err)
	}
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
