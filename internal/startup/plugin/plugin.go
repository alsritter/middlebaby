package plugin

import (
	"io"
	"os"
	"os/exec"

	"alsritter.icu/middlebaby/internal/proto"
)

type Serve interface {
	ServeRun() error
}

type Plugin struct {
	ProtoDescriptor      ProtoDescriptor
	AddArgs              AddArgs
	TargetAppRunBefore   TargetAppRunBefore
	TargetAppRunAfter    TargetAppRunAfter
	TargetAppStdOutput   TargetAppStdOutput
	TargetAppErrorOutput TargetAppErrorOutput
	ProcessExit          ProcessExit
	AllServerRunAfter    AllServerRunAfter

	// 启动的服务
	Server Serve
}

type Option func(*Plugin)

// BuildPlugin 构建一个插件
func BuildPlugin(server Serve, callbacks ...Option) Plugin {
	plg := defaultPlugin
	plg.Server = server
	for _, callback := range callbacks {
		callback(&plg)
	}
	return plg
}

var defaultPlugin = Plugin{
	AddArgs:              func() []string { return []string{} },
	TargetAppRunBefore:   func(command *exec.Cmd) {},
	TargetAppRunAfter:    func(command *exec.Cmd) {},
	TargetAppStdOutput:   func(reader io.ReadCloser) { _, _ = io.Copy(os.Stdout, reader) },
	TargetAppErrorOutput: func(reader io.ReadCloser) { _, _ = io.Copy(os.Stderr, reader) },
	ProtoDescriptor:      func(proto.Descriptor) {},
	ProcessExit:          func() { os.Exit(0) },
	AllServerRunAfter:    func() {},
}

// ProtoDescriptor Is executed when the proto file is loaded
type ProtoDescriptor func(proto.Descriptor)

// AddArgs
type AddArgs func() []string

// TargetAppStdOutput
type TargetAppStdOutput func(reader io.ReadCloser)

// TargetAppErrorOutput
type TargetAppErrorOutput func(reader io.ReadCloser)

// TargetAppRunBefore
type TargetAppRunBefore func(command *exec.Cmd)

// TargetAppRunAfter
type TargetAppRunAfter func(command *exec.Cmd)

// AllServerRunAfter
type AllServerRunAfter func()

// ProcessExit
type ProcessExit func()
