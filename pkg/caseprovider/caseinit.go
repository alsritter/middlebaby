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

package caseprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/file"
	"github.com/flynn/json5"
	"github.com/radovskyb/watcher"
)

// loading task server files and watcher these files modification.
func (b *basicProvider) init() error {
	b.loadGlobalMock()

	if err := b.loadFilePaths(); err != nil {
		return err
	}

	if err := b.loadCaseFiles(); err != nil {
		return err
	}

	if b.cfg.WatchCases {
		if err := b.watchCaseFiles(); err != nil {
			return err
		}
	}

	if b.cfg.WatchMock {
		if err := b.watchMockFiles(); err != nil {
			return err
		}
	}

	return nil
}

func (b *basicProvider) loadFilePaths() error {
	var (
		exists    = make(map[string]struct{})
		dirExists = make(map[string]struct{})
	)

	// find the absolute file path in CaseFiles.
	for _, filePath := range b.cfg.CaseFiles {
		// filePath could be './dir/**/json'
		matches, err := filepath.Glob(filePath)
		if err != nil {
			return fmt.Errorf("find file %s error: %v", filePath, err)
		}

		// real file path
		for _, matchPath := range matches {
			absFilePath, err := filepath.Abs(matchPath)
			if err != nil {
				return fmt.Errorf("get file %s absolute path error: %v", filePath, err)
			}

			// if exist, skip.
			if _, ok := exists[absFilePath]; !ok {
				exists[absFilePath] = struct{}{}
				// check file suffix
				if strings.HasSuffix(absFilePath, b.cfg.TaskFileSuffix) {
					b.taskFiles = append(b.taskFiles, absFilePath)
				}
			}

			// find all directory files.
			dirPath := filepath.Dir(absFilePath)
			absDirPath, err := filepath.Abs(dirPath)
			if err != nil {
				return fmt.Errorf("get directory %s absolute path error: %v", dirPath, err)
			}

			// if exist, skip.
			if _, ok := dirExists[absDirPath]; !ok {
				b.taskDirs = append(b.taskDirs, absDirPath)
				dirExists[absDirPath] = struct{}{}
			}
		}
	}

	return nil
}

// read all case files
func (b *basicProvider) loadCaseFiles() error {
	var total int
	b.mux.Lock()
	defer b.mux.Unlock()

	for _, file := range b.taskFiles {
		fb, err := ioutil.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return fmt.Errorf("read file: %s error: %v", file, err)
		}

		if err != nil {
			b.Error(nil, "gets the taskserver file %s service type error: %v \n", file, err)
			continue
		}

		var t ItfTask
		if err := json5.Unmarshal(fb, &t); err != nil {
			return fmt.Errorf("serialization %s file error: %v", file, err)
		}

		if err := b.checkItfInfo(*t.TaskInfo); err != nil {
			return err
		}

		// check case name
		for _, e := range t.Cases {
			if err := b.checkCaseInfo(e, *t.TaskInfo); err != nil {
				return err
			}
			b.mockCases[e.Name] = append(b.mockCases[e.Name], e.Mocks...)
		}

		total += len(t.Cases)
		if _, ok := b.taskInterface[t.ServiceName]; ok {
			return fmt.Errorf("the serviceName is exists in multiple files, duplicated service name: [%s]", t.ServiceName)
		} else {
			b.taskInterface[t.ServiceName] = &t
		}

		fileInfo, _ := os.Stat(file)

		b.taskWithFileInfo[t.ServiceName] = &ItfTaskWithFileInfo{
			Dirpath:      path.Dir(file),
			Filename:     fileInfo.Name(),
			ModifiedTime: fileInfo.ModTime(),
			ItfTask:      b.taskInterface[t.ServiceName],
		}

		// add interface mocks case
		b.mockCases[t.ServiceName] = append(b.mockCases[t.ServiceName], t.Mocks...)
	}

	b.Info(nil, "loading all case, total: %d", total)
	return nil
}

// Check whether the file is correct.
func (b *basicProvider) checkItfInfo(info TaskInfo) error {
	if info.ServiceName == globalCaseID {
		return fmt.Errorf("interface name cannot be %s", globalCaseID)
	}

	if info.Protocol == ProtocolGRPC && info.ServiceProtoFile == "" {
		return fmt.Errorf("grpc request proto file path cannot be empty")
	}

	return nil
}

func (b *basicProvider) checkCaseInfo(e *CaseTask, info TaskInfo) error {
	if e.Name == info.ServiceName {
		return fmt.Errorf("case name cannot be the same as interface name %s", e.Name)
	}

	if e.Name == globalCaseID {
		return fmt.Errorf("case name cannot be %s", globalCaseID)
	}

	return nil
}

// Listen for changes to the task server file
func (b *basicProvider) watchCaseFiles() error {
	var paths []string
	paths = append(paths, b.taskFiles...)
	paths = append(paths, b.taskDirs...)

	w, err := file.InitializeWatcher(paths...)
	if err != nil {
		return fmt.Errorf("failed to start test case description file listening %v", err)
	}

	file.AttachWatcher(w, func(event watcher.Event) {
		b.Trace(nil, "listening file event is triggered: %v", event)
		// If it is a file creation event, It is added to the listener
		if event.Op == watcher.Create {
			if strings.HasSuffix(event.Name(), b.cfg.TaskFileSuffix) {
				fi, err := os.Stat(event.Name())
				// if you created a directory.
				if err == nil && fi.IsDir() {
					if err := w.AddRecursive(event.Name()); err != nil {
						b.Error(nil, "Add test case directory listening %s :%s \n", event.Name, err.Error())
					}
					return
				} else {
					// if exist, skip.
					for _, f := range b.taskFiles {
						if f == event.Name() {
							return
						}
					}

					b.taskFiles = append(b.taskFiles, event.Name())
				}
			}
		}

		// clear all cases
		b.clearAllData()

		// FIXME: Global loading is not required here
		if err := b.loadCaseFiles(); err != nil {
			b.Error(nil, "Failed to re-read task server file error: ", err)
		}

		if event.Op != watcher.Remove {
			// TODO: reload listening files
			_ = w.AddRecursive(event.Name())
		}
	})
	return nil
}

func (b *basicProvider) loadGlobalMock() {
	b.clearGlobalMock()
	for _, filePath := range b.cfg.MockFiles {
		b.loadSingleImposter(filePath)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (b *basicProvider) watchMockFiles() error {
	w, err := file.InitializeWatcher(b.cfg.MockFiles...)
	if err != nil {
		return fmt.Errorf("initialize watcher failed: %v", err)
	}

	// FIXME: Global loading is not required here
	file.AttachWatcher(w, func(evn watcher.Event) {
		b.clearGlobalMock()
		b.loadGlobalMock()
		// b.loadSingleImposter(evn.Path)
	})

	return nil
}

// loading single case file to imposter
func (b *basicProvider) loadSingleImposter(filePath string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if !filepath.IsAbs(filePath) {
		if fp, err := filepath.Abs(filePath); err != nil {
			b.Error(nil, "to absolute representation path err: %s", err)
			return
		} else {
			filePath = fp
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		b.Error(nil, "%v: error trying to read config file: %s", err, filePath)
	}

	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []*interact.ImposterCase
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		b.Error(nil, "%v: error while unmarshal configFile file %s", err, filePath)
	}

	b.mockCases[globalCaseID] = append(b.mockCases[globalCaseID], imposter...)
}

func (b *basicProvider) clearAllData() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.taskInterface = make(map[string]*ItfTask)
	b.mockCases = make(map[string][]*interact.ImposterCase)
	b.taskWithFileInfo = make(map[string]*ItfTaskWithFileInfo)
}

func (b *basicProvider) clearGlobalMock() {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.mockCases, globalCaseID)
}
