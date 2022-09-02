package caseprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/file"
	"github.com/flynn/json5"
	"github.com/radovskyb/watcher"
)

// loading task server files and watcher these files modification.
func (b *basicProvider) init() error {
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
			return fmt.Errorf("find file %s error: %w", filePath, err)
		}

		// real file path
		for _, matchPath := range matches {
			absFilePath, err := filepath.Abs(matchPath)
			if err != nil {
				return fmt.Errorf("get file %s absolute path error: %w", filePath, err)
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
				return fmt.Errorf("get directory %s absolute path error: %w", dirPath, err)
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
			return fmt.Errorf("read file: %s error: %w", file, err)
		}

		if err != nil {
			b.Error(nil, "gets the taskserver file %s service type error: %w \n", file, err)
			continue
		}

		var t InterfaceTask
		if err := json5.Unmarshal(fb, &t); err != nil {
			return fmt.Errorf("serialization %s file error: %w", file, err)
		}

		if t.ServiceName == globalCaseID {
			return fmt.Errorf("interface name cannot be %s", globalCaseID)
		}

		// check case name
		for _, e := range t.Cases {
			if e.Name == t.ServiceName {
				return fmt.Errorf("case name cannot be the same as interface name %s", e.Name)
			}

			if e.Name == globalCaseID {
				return fmt.Errorf("case name cannot be %s", globalCaseID)
			}

			b.mockCases[e.Name] = append(b.mockCases[e.Name], e.Mocks...)
		}

		total += len(t.Cases)
		// if exist, add case to here..
		if old, ok := b.taskInterface[t.ServiceName]; !ok {
			// Check that the use case is duplicated
			hash := make(map[string]bool)
			for _, e := range old.Cases {
				hash[e.Name] = true
			}

			for _, e := range t.Cases {
				if hash[e.Name] {
					return fmt.Errorf("case name is duplicated %s", e.Name)
				}
			}

			// add case
			old.Cases = append(old.Cases, t.Cases...)
		} else {
			b.taskInterface[t.ServiceName] = &t
		}

		// add interface mocks case
		b.mockCases[t.ServiceName] = append(b.mockCases[t.ServiceName], t.Mocks...)
	}

	b.Info(nil, "loading all case, total: %d", total)
	return nil
}

// Listen for changes to the task server file
func (b *basicProvider) watchCaseFiles() error {
	var paths []string
	paths = append(paths, b.taskFiles...)
	paths = append(paths, b.taskDirs...)

	w, err := file.InitializeWatcher(paths...)
	if err != nil {
		return fmt.Errorf("failed to start test case description file listening %w", err)
	}

	file.AttachWatcher(w, func(event watcher.Event) {
		b.Trace(nil, "listening file event is triggered: ", event)
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

func (b *basicProvider) loadImposter() {
	b.clearGlobalMock()
	for _, filePath := range b.cfg.MockFiles {
		b.loadSingleImposter(filePath)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (b *basicProvider) watchMockFiles() error {
	w, err := file.InitializeWatcher(b.cfg.MockFiles...)
	if err != nil {
		return fmt.Errorf("initialize watcher failed: %w", err)
	}

	// FIXME: Global loading is not required here
	file.AttachWatcher(w, func(evn watcher.Event) {
		b.clearGlobalMock()
		b.loadImposter()
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
		b.Error(nil, "%w: error trying to read config file: %s", err, filePath)
	}

	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []*interact.ImposterCase
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		b.Error(nil, "%w: error while unmarshal configFile file %s", err, filePath)
	}

	b.mockCases[globalCaseID] = append(b.mockCases[globalCaseID], imposter...)
}

func (b *basicProvider) clearAllData() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.taskInterface = make(map[string]*InterfaceTask)
	b.mockCases = make(map[string][]*interact.ImposterCase)
}

func (b *basicProvider) clearGlobalMock() {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.mockCases, globalCaseID)
}
