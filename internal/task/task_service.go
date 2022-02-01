package task

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"alsritter.icu/middlebaby/internal/file/common"
	"alsritter.icu/middlebaby/internal/file/task_file"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"alsritter.icu/middlebaby/internal/startup/plugin"
	"github.com/flynn/json5"
	"github.com/radovskyb/watcher"
)

type TestCaseType = string

const (
	TestCaseTypeHTTP TestCaseType = "http"
	TestCaseTypeGRpc TestCaseType = "grpc"
)

// represents a TaskName and all cases under it.
type TaskCaseTree struct {
	InterfaceName string   // Task Name(Interface Name)
	CaseList      []string // Case Names
}

// grpc or http runner interface.
type ITaskRunner interface {
	// Run execution test case.
	Run(caseName string, env plugin.Env, mockCenter proxy.MockCenter, runner Runner) error
	// Get All Task and the Task's Cases
	GetTaskCaseTree() []*TaskCaseTree
}

type TaskService struct {
	// all test case files. (file absolute path)
	taskFiles []string
	// all test case directory. (absolute path)
	taskDirs []string
	// provides an interface for use case execution.
	runner Runner
	// mock center
	mockCenter proxy.MockCenter
	// configuration information required by the service.
	env plugin.Env
	// save all task runner
	taskRunners map[TestCaseType]ITaskRunner
}

// return a TaskService
func NewTaskService(env plugin.Env, mockCenter proxy.MockCenter, runner Runner) (*TaskService, error) {
	ts := new(TaskService)
	ts.env = env
	ts.taskRunners = make(map[TestCaseType]ITaskRunner)
	ts.mockCenter = mockCenter
	ts.runner = runner

	return ts, ts.init()
}

// loading task files and watcher these files modification.
func (t *TaskService) init() error {
	// find the absolute file path in cfgFilePaths.
	for _, filePath := range t.env.GetConfig().CaseFiles {
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
	if len(t.taskFiles) == 0 && t.env.GetConfig().TaskFileSuffix == "" {
		return fmt.Errorf("no test case files were found")
	}

	// because maybe the cfgFilePaths is the directory path, so we need to find the directory path.
	if t.env.GetConfig().TaskFileSuffix != "" {
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

	if t.env.GetConfig().Watcher {
		if err := t.watchFiles(); err != nil {
			return err
		}
	}

	return nil
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
			log.Errorf("gets the task file %s service type error: %w \n", file, err)
			continue
		}

		if testCaseType == TestCaseTypeHTTP {
			if tk, err := t.unmarshalHttp(fb); err != nil {
				log.Errorf("serialization file: %s error: %w \n", file, err)
				continue
			} else {
				httpTaskList = append(httpTaskList, tk)
			}
		} else if testCaseType == TestCaseTypeGRpc {
			if tk, err := t.unmarshalGrpc(fb); err != nil {
				log.Errorf("serialization file: %s error: %w \n", file, err)
				continue
			} else {
				grpcTaskList = append(grpcTaskList, tk)
			}

		} else {
			log.Error("unknown service type: ", testCaseType)
			continue
		}
	}

	t.taskRunners[TestCaseTypeGRpc] = newGRpcTaskRunner(grpcTaskList)
	t.taskRunners[TestCaseTypeHTTP] = newHttpTaskRunner(httpTaskList)
	log.Trace("loading all task file.")
	return nil
}

// unmarshal http task file.
func (t *TaskService) unmarshalHttp(testCaseFileByte []byte) (*task_file.HttpTask, error) {
	var httpTask task_file.HttpTask
	if err := json5.Unmarshal(testCaseFileByte, &httpTask); err != nil {
		return nil, fmt.Errorf("serialization HTTP task file error: %w", err)
	}

	log.Tracef("%#v \n", httpTask)

	return &httpTask, nil
}

// unmarshal grpc task file.
func (t *TaskService) unmarshalGrpc(testCaseFileByte []byte) (*task_file.GRpcTask, error) {
	var grpcTask task_file.GRpcTask
	if err := json5.Unmarshal(testCaseFileByte, &grpcTask); err != nil {
		return nil, fmt.Errorf("serialization GRPC task file error: %w", err)
	}
	return &grpcTask, nil
}

// get task file type.
func (t *TaskService) getTestCaseType(fileByte []byte) (TestCaseType, error) {
	type ServiceType struct {
		ServiceType string `json:"serviceType"`
	}

	// there can be multiple tasks in a file, so needs use slice save they.
	var serviceType ServiceType
	if err := json5.Unmarshal(fileByte, &serviceType); err != nil {
		return "", fmt.Errorf("unmarshal task file error: %#v", err)
	}

	var uniqType = make(map[TestCaseType]bool)
	var testCaseType TestCaseType

	testCaseType = TestCaseType(strings.ToLower(serviceType.ServiceType))
	switch testCaseType {
	case TestCaseTypeHTTP, TestCaseTypeGRpc:
	default:
		testCaseType = TestCaseTypeGRpc
	}
	uniqType[testCaseType] = true

	// there is no service in this file
	if testCaseType == "" {
		return "", fmt.Errorf("the task file type does not exist")
	}

	return testCaseType, nil
}

// Listen for changes to the task file
func (t *TaskService) watchFiles() error {
	var paths []string
	paths = append(paths, t.taskFiles...)
	paths = append(paths, t.taskDirs...)
	w, err := common.InitializeWatcher(paths...)
	if err != nil {
		return fmt.Errorf("failed to start test case description file listening %w", err)
	}

	common.AttachWatcher(w, func(event watcher.Event) {
		log.Trace("listening file event is triggered: ", event)
		// If it is a file creation event, It is added to the listener
		if event.Op == watcher.Create {
			if t.env.GetConfig().TaskFileSuffix != "" && strings.HasSuffix(event.Name(), t.env.GetConfig().TaskFileSuffix) {
				fi, err := os.Stat(event.Name())
				// if you created a directory.
				if err == nil && fi.IsDir() {
					if err := w.AddRecursive(event.Name()); err != nil {
						log.Errorf("Add test case directory listening %s :%s \n", event.Name, err.Error())
					}
					return
				} else {
					t.addTestCaseFile(event.Name())
				}
			}
		}

		// otherwise, clear all task files and reload the listener.
		t.removeAllServices()
		if err := t.readTaskCaseFiles(); err != nil {
			log.Error("Failed to re-read task file error: ", err)
		}

		if event.Op != watcher.Remove {
			// TODO: reload listening files
			_ = w.AddRecursive(event.Name())
		}
	})
	return nil
}

func (t *TaskService) removeAllServices() {
	t.taskRunners = make(map[TestCaseType]ITaskRunner)
}

// Run specified case
func (t *TaskService) Run(caseType TestCaseType, caseName string, mustRunTearDown *bool) error {
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

	return caseRunner.Run(caseName, t.env, t.mockCenter, t.runner)
}

// GetAllTestCase
func (t *TaskService) GetAllTestCase() map[TestCaseType]ITaskRunner {
	return t.taskRunners
}
