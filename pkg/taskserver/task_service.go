package taskserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/runner"
	"github.com/alsritter/middlebaby/pkg/taskserver/grpc_runner"
	"github.com/alsritter/middlebaby/pkg/taskserver/http_runner"
	"github.com/alsritter/middlebaby/pkg/util/file"
	"github.com/alsritter/middlebaby/pkg/util/logger"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/taskserver/task_file"
	"github.com/flynn/json5"
	"github.com/radovskyb/watcher"
)

const (
	TestCaseTypeHTTP task_file.TestCaseType = "http"
	TestCaseTypeGRpc task_file.TestCaseType = "grpc"
)

type Config struct {
	CaseFiles       []string `yaml:"caseFiles"`
	TaskFileSuffix  string   `yaml:"taskFileSuffix"` // the default test case suffix name. example: ".case.json"
	WatcherCases    bool     `yaml:"watcherCases"`
	MustRunTearDown bool     `yaml:"mustRunTearDown"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

// var _ hello_proto.HelloServiceServer = (*TaskService)(nil)

type Provider interface {
	GetAllTestCase() map[task_file.TestCaseType]runner.ITaskRunner
}

type TaskService struct {
	// all test case files. (file absolute path)
	taskFiles []string
	// all test case directory. (absolute path)
	taskDirs []string
	// provides an interface for use case execution.
	runner runner.Runner
	// mock center
	mockCenter apimanager.MockCaseCenter
	// configuration information required by the service.
	cfg *Config

	// save all task server runner
	taskRunners map[task_file.TestCaseType]runner.ITaskRunner
	log         logger.Logger
}

// New return a TaskService
func New(log logger.Logger, cfg *Config, mockCenter apimanager.MockCaseCenter, r runner.Runner) (*TaskService, error) {
	ts := &TaskService{
		runner:      r,
		mockCenter:  mockCenter,
		cfg:         cfg,
		taskRunners: make(map[task_file.TestCaseType]runner.ITaskRunner),
		log:         log.NewLogger("task"),
	}
	return ts, ts.init()
}

// loading task server files and watcher these files modification.
func (t *TaskService) init() error {
	// find the absolute file path in cfgFilePaths.
	for _, filePath := range t.cfg.CaseFiles {
		matches, err := filepath.Glob(filePath)
		if err != nil {
			return fmt.Errorf("find file %s error: %w", filePath, err)
		}

		for _, matchFile := range matches {
			absFilePath, err := filepath.Abs(matchFile)
			if err != nil {
				return fmt.Errorf("get file %s absolute path error: %w", filePath, err)
			}

			t.addTestCaseFile(absFilePath)
		}
	}

	// no test case and no file suffix set
	if len(t.taskFiles) == 0 && t.cfg.TaskFileSuffix == "" {
		return fmt.Errorf("no test case files were found")
	}

	// because maybe the cfgFilePaths is the directory path, so we need to find the directory path.
	if t.cfg.TaskFileSuffix != "" {
		// If a directory exists, find all directory files.
		for _, filePath := range t.taskFiles {
			dirPath := filepath.Dir(filePath)
			absDirPath, err := filepath.Abs(dirPath)
			if err != nil {
				return fmt.Errorf("get directory%s absolute path error: %w", dirPath, err)
			}
			t.addTestCaseDir(absDirPath)
		}
	}

	if err := t.readTaskCaseFiles(); err != nil {
		return err
	}

	if t.cfg.WatcherCases {
		if err := t.watchFiles(); err != nil {
			return err
		}
	}

	return nil
}

func (t TaskService) Start() error {

	return nil
}

func (t TaskService) Close() error {
	return nil
}

// Run specified case
func (t *TaskService) Run(caseType task_file.TestCaseType, caseName string) error {
	caseRunner, ok := t.taskRunners[caseType]
	if !ok {
		return fmt.Errorf("there are no test cases of this type: %s", caseType)
	}

	if t.runner == nil {
		return fmt.Errorf("runner cannot be empty")
	}

	if t.mockCenter == nil {
		return fmt.Errorf("mockCenter cannot be empty")
	}

	return caseRunner.Run(caseName, t.mockCenter, t.runner)
}

func (t *TaskService) GetAllTestCase() map[task_file.TestCaseType]runner.ITaskRunner {
	return t.taskRunners
}

// add file path
func (t *TaskService) addTestCaseFile(filePath string) {
	// if exist, skip.
	for _, f := range t.taskFiles {
		if f == filePath {
			return
		}
	}

	t.taskFiles = append(t.taskFiles, filePath)
}

// add directory path
func (t *TaskService) addTestCaseDir(dir string) {
	// if exist, skip.
	for _, d := range t.taskDirs {
		if d == dir {
			return
		}
	}

	t.taskDirs = append(t.taskDirs, dir)
}

// read all case files
func (t *TaskService) readTaskCaseFiles() error {
	var httpTaskList []*task_file.HttpTask
	var grpcTaskList []*task_file.GRpcTask

	for _, file := range t.taskFiles {
		fb, err := ioutil.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return fmt.Errorf("read file: %s error: %w", file, err)
		}

		testCaseType, err := t.getTestCaseType(fb)
		if err != nil {
			t.log.Error(nil, "gets the taskserver file %s service type error: %w \n", file, err)
			continue
		}

		if testCaseType == TestCaseTypeHTTP {
			if tk, err := t.unmarshalHttp(fb); err != nil {
				t.log.Error(nil, "serialization file: %s error: %w \n", file, err)
				continue
			} else {
				httpTaskList = append(httpTaskList, tk)
			}
		} else if testCaseType == TestCaseTypeGRpc {
			if tk, err := t.unmarshalGrpc(fb); err != nil {
				t.log.Error(nil, "serialization file: %s error: %w \n", file, err)
				continue
			} else {
				grpcTaskList = append(grpcTaskList, tk)
			}

		} else {
			t.log.Error(nil, "unknown service type: ", testCaseType)
			continue
		}
	}

	t.taskRunners[TestCaseTypeGRpc] = grpc_runner.New(grpcTaskList, t.log)
	t.taskRunners[TestCaseTypeHTTP] = http_runner.New(httpTaskList, t.log)
	t.log.Info(nil, "loading all task server file, total: %d", len(grpcTaskList)+len(httpTaskList))
	return nil
}

// unmarshal http task server file.
func (t *TaskService) unmarshalHttp(testCaseFileByte []byte) (*task_file.HttpTask, error) {
	var httpTask task_file.HttpTask
	if err := json5.Unmarshal(testCaseFileByte, &httpTask); err != nil {
		return nil, fmt.Errorf("serialization HTTP taskserver file error: %w", err)
	}
	return &httpTask, nil
}

// unmarshal grpc task server file.
func (t *TaskService) unmarshalGrpc(testCaseFileByte []byte) (*task_file.GRpcTask, error) {
	var grpcTask task_file.GRpcTask
	if err := json5.Unmarshal(testCaseFileByte, &grpcTask); err != nil {
		return nil, fmt.Errorf("serialization GRPC taskserver file error: %w", err)
	}
	return &grpcTask, nil
}

// get task server file type.
func (t *TaskService) getTestCaseType(fileByte []byte) (task_file.TestCaseType, error) {
	type ServiceType struct {
		ServiceType string `json:"serviceType"`
	}

	// there can be multiple tasks in a file, so needs use slice save they.
	var serviceType ServiceType
	if err := json5.Unmarshal(fileByte, &serviceType); err != nil {
		return "", fmt.Errorf("unmarshal taskserver file error: %#v", err)
	}

	var uniqType = make(map[task_file.TestCaseType]bool)
	var testCaseType task_file.TestCaseType

	testCaseType = strings.ToLower(serviceType.ServiceType)
	switch testCaseType {
	case TestCaseTypeHTTP, TestCaseTypeGRpc:
	default:
		testCaseType = TestCaseTypeGRpc
	}
	uniqType[testCaseType] = true

	// there is no service in this file
	if testCaseType == "" {
		return "", fmt.Errorf("the taskserver file type does not exist")
	}

	return testCaseType, nil
}

// Listen for changes to the task server file
func (t *TaskService) watchFiles() error {
	var paths []string
	paths = append(paths, t.taskFiles...)
	paths = append(paths, t.taskDirs...)
	w, err := file.InitializeWatcher(paths...)
	if err != nil {
		return fmt.Errorf("failed to start test case description file listening %w", err)
	}

	file.AttachWatcher(w, func(event watcher.Event) {
		t.log.Trace(nil, "listening file event is triggered: ", event)
		// If it is a file creation event, It is added to the listener
		if event.Op == watcher.Create {
			if t.cfg.TaskFileSuffix != "" && strings.HasSuffix(event.Name(), t.cfg.TaskFileSuffix) {
				fi, err := os.Stat(event.Name())
				// if you created a directory.
				if err == nil && fi.IsDir() {
					if err := w.AddRecursive(event.Name()); err != nil {
						t.log.Error(nil, "Add test case directory listening %s :%s \n", event.Name, err.Error())
					}
					return
				} else {
					t.addTestCaseFile(event.Name())
				}
			}
		}

		// otherwise, clear all taskserver files and reload the listener.
		t.removeAllServices()
		if err := t.readTaskCaseFiles(); err != nil {
			t.log.Error(nil, "Failed to re-read task server file error: ", err)
		}

		if event.Op != watcher.Remove {
			// TODO: reload listening files
			_ = w.AddRecursive(event.Name())
		}
	})
	return nil
}

func (t *TaskService) removeAllServices() {
	t.taskRunners = make(map[task_file.TestCaseType]runner.ITaskRunner)
}
