package protomanager

import (
	"os"
	"sync"
	"testing"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/synchronization"
	"golang.org/x/net/context"
)

func TestManager_GetMethod(t *testing.T) {
	var (
		wg          *sync.WaitGroup
		ctx, cancel = context.WithCancel(context.Background())
		clog        = logger.NewDefault("test")
	)

	basePath, _ := os.Getwd()
	pms, err := New(&Config{
		ProtoDir:         basePath,
		ProtoImportPaths: []string{"temporary/alsritter/protobuf-examples"},
		Synchronization: &synchronization.Config{
			Enable:     true,
			StorageDir: "temporary",
			Repository: []*synchronization.Repository{{Address: "git@github.com:alsritter/protobuf-examples.git", Branch: "main"}},
		},
	}, clog)
	if err != nil {
		t.Error(err)
	}

	err = pms.Start(ctx, cancel, wg)
	if err != nil {
		t.Error(err)
	}

	d, ext := pms.GetMethod("/hello.Hello/SayHello")
	if ext {
		t.Logf("查询到的服务全地址为：%#v", d.GetFullyQualifiedName())
	} else {
		t.Error("不存在")
	}
}
