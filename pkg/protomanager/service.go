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

package protomanager

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/grpcurl"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
	"github.com/alsritter/middlebaby/pkg/util/synchronization"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/spf13/pflag"
)

// Config defines the config structure
type Config struct {
	ProtoImportPaths []string
	SyncGitManger    *synchronization.Config `yaml:"sync"`
}

func NewConfig() *Config {
	return &Config{
		ProtoImportPaths: []string{},
		SyncGitManger:    synchronization.NewConfig(),
	}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
}

// Provider is used to read, parse and manage Proto files
type Provider interface {
	Start(ctx *mbcontext.Context) error
	// GetMethod is used to get descriptor of specified grpc path
	GetMethod(name string) (*desc.MethodDescriptor, bool)
	// get proto importPaths
	GetImportPaths() []string
}

// Manager is the implement of Provider
type Manager struct {
	cfg *Config

	// map[name]*desc.MethodDescriptor
	methods         *sync.Map
	methodsLock     sync.Mutex
	synchronization *synchronization.Service

	logger.Logger
}

// New is used to init service
func New(log logger.Logger, cfg *Config) (Provider, error) {
	service := &Manager{
		cfg:     cfg,
		methods: &sync.Map{},
		Logger:  log.NewLogger("proto"),
	}
	if cfg.SyncGitManger.Enable {
		s, err := synchronization.New(cfg.SyncGitManger, log)
		if err != nil {
			return nil, err
		}
		service.synchronization = s
		if err := service.synchronizeProto(context.Background(), true); err != nil {
			return nil, err
		}
	} else {
		if err := service.loadProto(); err != nil {
			return nil, err
		}
	}
	return service, nil
}

// GetImportPaths implements Provider
func (m *Manager) GetImportPaths() []string {
	return m.cfg.ProtoImportPaths
}

// GetMethod is used to get descriptor of specified grpc path
func (s *Manager) GetMethod(name string) (*desc.MethodDescriptor, bool) {
	s.methodsLock.Lock()
	method := s.methods
	s.methodsLock.Unlock()
	val, ok := method.Load(name)
	if !ok {
		return nil, false
	}
	return val.(*desc.MethodDescriptor), true
}

func (s *Manager) Start(ctx *mbcontext.Context) error {
	if err := s.startSynchronization(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Manager) startSynchronization(ctx *mbcontext.Context) error {
	if !s.cfg.SyncGitManger.Enable {
		return nil
	}
	util.StartServiceAsync(ctx, s.Logger, func() error {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Info(nil, "start to synchronize proto")
				if err := s.synchronizeProto(ctx, false); err != nil {
					s.Error(map[string]interface{}{
						"err": err,
					}, "failed to synchronize proto")
				}
				s.Info(nil, "synchronize proto finished")
			case <-ctx.Done():
				return nil
			}
		}
	}, func() error {
		return nil
	})
	return nil
}

func (s *Manager) synchronizeProto(ctx context.Context, force bool) error {
	var shouldReload bool
	_ = s.synchronization.Synchronize(ctx, func(repository string, updated bool, err error) error {
		if err != nil {
			s.Warn(map[string]interface{}{
				"error":      err,
				"repository": repository,
			}, "failed to synchronize repository")
			return nil
		}
		s.Info(map[string]interface{}{
			"repository": repository,
			"updated":    updated,
		}, "repository synchronized")
		if updated {
			shouldReload = true
		}
		return nil
	})
	if shouldReload || force {
		if err := s.loadProto(); err != nil {
			return err
		}
	}
	return nil
}

// 加载 Proto 文件
func (s *Manager) loadProto() error {
	var (
		methods sync.Map
		count   int
	)

	importPaths := s.cfg.ProtoImportPaths
	for _, importProtoPath := range importPaths {
		if err := filepath.Walk(importProtoPath, func(path string, info os.FileInfo, e error) error {
			if info == nil || info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), ".proto") {
				return nil
			}
			// fileNames = append(fileNames, path)
			relPath, err := filepath.Rel(importProtoPath, path)
			if err != nil {
				return err
			}
			parser := protoparse.Parser{
				ImportPaths:           importPaths,
				InferImportPaths:      len(importPaths) == 0,
				IncludeSourceCodeInfo: true,
				Accessor:              grpcurl.Accessor,
			}
			fds, err := parser.ParseFiles(relPath)
			if err != nil {
				s.Error(nil, "failed to parse file: %s", err)
				return nil
			}
			for _, fd := range fds {
				for _, service := range fd.GetServices() {
					for _, method := range service.GetMethods() {
						name := GetPathByFullyQualifiedName(method.GetFullyQualifiedName())
						s.Info(map[string]interface{}{
							"name": name,
						}, "proto loaded")
						_, loaded := methods.LoadOrStore(name, method)
						if loaded {
							s.Trace(map[string]interface{}{
								"name":  name,
								"error": "method already exists",
							}, "failed to load method")
							continue
						}
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	s.Info(map[string]interface{}{
		"total":            count,
		"importProtoPaths": importPaths,
	}, "methods loaded")

	s.methodsLock.Lock()
	s.methods = &methods
	s.methodsLock.Unlock()
	return nil
}

// GetPathByFullyQualifiedName is used to get the grpc path of specified fully qualified name
func GetPathByFullyQualifiedName(name string) string {
	raw := []byte(name)
	if i := bytes.LastIndexByte(raw, '.'); i > 0 {
		raw[i] = '/'
	}
	return string(append([]byte{'/'}, raw...))
}
