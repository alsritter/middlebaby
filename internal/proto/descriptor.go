package proto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"alsritter.icu/middlebaby/internal/log"
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
	log.Tracef("parse the grpc mock file %+v \n", fileNames)
	fileNames, err := protoparse.ResolveFilenames(importProtoPaths, fileNames...)
	if err != nil {
		return nil, err
	}

	log.Tracef("after parsing the grpc mock file %+v \n", fileNames)
	p := protoparse.Parser{
		ImportPaths:           importProtoPaths,
		InferImportPaths:      len(importProtoPaths) == 0,
		IncludeSourceCodeInfo: true,
	}

	var fds []*desc.FileDescriptor
	for _, filename := range fileNames {
		fd, err := p.ParseFiles(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to parse files: %v", err)
		}

		log.Tracef("Parse file:%s %d %s %+v \n", filename, len(fd), fd[0].GetPackage(), fd[0].GetServices())
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
