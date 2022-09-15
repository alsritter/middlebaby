/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package targetprocess

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/alsritter/middlebaby/pkg/mockserver"
	"github.com/alsritter/middlebaby/pkg/types/target"

	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
)

type Config struct {
	AppPath  string       `yaml:"appPath"`
	Env      []target.Env `json:"env"`
	mockPort int          `yaml:"-" json:"-"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.AppPath == "" {
		return fmt.Errorf("the target application cannot be empty")
	}

	// Check if your app file exists
	if _, err := os.Stat(c.AppPath); err != nil {
		return fmt.Errorf("check if your target application file exists [%s], error: [%v]", c.AppPath, err)
	}

	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.AppPath, prefix+"target.path", c.AppPath, "target application address")
}

// Provider defines the target process interface
type Provider interface {
	Start(ctx *mbcontext.Context) error
	GetRuntimeInfo() *target.RuntimeInfo
}

type TargetProcess struct {
	cfg     *Config
	command *exec.Cmd
	logger.Logger
	cwd   string    // current working directory
	birth time.Time // service start time
}

func New(log logger.Logger, cfg *Config, mock mockserver.Provider) Provider {
	cfg.mockPort = mock.GetPort()
	return &TargetProcess{
		cfg:    cfg,
		Logger: log.NewLogger("target"),
	}
}

// GetRuntimeInfo implements Provider
func (t *TargetProcess) GetRuntimeInfo() *target.RuntimeInfo {
	return &target.RuntimeInfo{
		StartTime:      t.birth,
		CWD:            t.cwd,
		GoroutineCount: runtime.NumGoroutine(),
		GOMAXPROCS:     runtime.GOMAXPROCS(0),
		GOGC:           os.Getenv("GOGC"),
		GODEBUG:        os.Getenv("GODEBUG"),
	}
}

// Start the service to be tested
func (t *TargetProcess) Start(ctx *mbcontext.Context) error {
	util.StartServiceAsync(ctx, t, func() error {
		if _, err := os.Stat(t.cfg.AppPath); err != nil {
			return fmt.Errorf("target app err: %v", err)
		}

		// record runtime info
		t.birth = time.Now()
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "<error retrieving current working directory>"
		}
		t.cwd = cwd

		for _, env := range t.cfg.Env {
			t.Debug(nil, "setting environment variables [%+v]", env)
			if err := os.Setenv(env.Name, env.Value); err != nil {
				t.Error(nil, "setting environment variable: [%+v] error: [%v]", env, err)
			}
		}

		// preparing to start service
		t.command = exec.Command(t.cfg.AppPath)
		port := t.cfg.mockPort
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

		// Remove here.
		time.Sleep(3 * time.Second) // wait for mock server running.

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
