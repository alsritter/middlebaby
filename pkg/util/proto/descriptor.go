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

package proto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alsritter/middlebaby/pkg/util/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

// Look for the proto file in that directory.
func DescriptorSourceFromProtoFiles(importProtoPaths []string) (Descriptor, error) {
	var fileNames []string
	for _, importProtoPath := range importProtoPaths {

		// Traverses all files in the specified directory
		if err := filepath.Walk(importProtoPath, func(path string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), ".proto") {
				return nil
			}

			fileNames = append(fileNames, path)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Append relative paths to avoid the file name duplication after resolve. If the file name is the same, the resolve BUG cannot be resolved
	fds, err := resolveFileNames(append([]string{"."}, importProtoPaths...), fileNames...)
	if err != nil {
		return nil, err
	}

	return &fileDescriptor{fds: fds}, nil
}

func resolveFileNames(importProtoPaths []string, fileNames ...string) ([]*desc.FileDescriptor, error) {
	// log.Tracef("parse the grpc mock file %+v \n", fileNames)
	fileNames, err := protoparse.ResolveFilenames(importProtoPaths, fileNames...)
	if err != nil {
		return nil, err
	}

	// log.Tracef("after parsing the grpc mock file %+v \n", fileNames)
	p := protoparse.Parser{
		ImportPaths:           importProtoPaths,
		InferImportPaths:      len(importProtoPaths) == 0,
		IncludeSourceCodeInfo: true,
		Accessor:              grpcurl.Accessor,
	}

	var fds []*desc.FileDescriptor
	for _, filename := range fileNames {
		fd, err := p.ParseFiles(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to parse files: %v", err)
		}

		// log.Tracef("Parse file:%s %d %s %+v \n", filename, len(fd), fd[0].GetPackage(), fd[0].GetServices())
		fds = append(fds, fd...)
	}
	return fds, nil
}

type Descriptor interface {
	// FindSymbol Find the description of the symbol
	FindSymbol(symbol string) (desc.Descriptor, error)
}

type fileDescriptor struct {
	fds []*desc.FileDescriptor
}

func (f *fileDescriptor) FindSymbol(symbol string) (desc.Descriptor, error) {
	for _, fd := range f.fds {
		if dsc := fd.FindSymbol(symbol); dsc != nil {
			return dsc, nil
		}
	}
	return nil, fmt.Errorf("cannot find symbol: %s", symbol)
}

func FindMethod(descriptor Descriptor, svcAndMethod string) (*desc.ServiceDescriptor, string, error) {
	svc, mth := parseSymbol(svcAndMethod)
	dsc, err := descriptor.FindSymbol(svc)
	if err != nil {
		err = fmt.Errorf("cannot find grpc method: [%v]", err)
		return nil, "", err
	}

	sd, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		err = fmt.Errorf("find gRpc method: [%s] but not a method type:desc.ServiceDescriptor", svc)
		return nil, "", err
	}
	return sd, mth, nil
}

func parseSymbol(svcAndMethod string) (string, string) {
	svcAndMethod = strings.TrimPrefix(svcAndMethod, "/")
	pos := strings.LastIndex(svcAndMethod, "/")
	if pos < 0 {
		pos = strings.LastIndex(svcAndMethod, ".")
		if pos < 0 {
			return "", ""
		}
	}
	return svcAndMethod[:pos], svcAndMethod[pos+1:]
}
