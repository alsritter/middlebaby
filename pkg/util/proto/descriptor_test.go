package proto

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jhump/protoreflect/desc"
)

var testProto = `
syntax = "proto3";
package hello;

option go_package = "./;hello";

service HelloService {
    rpc SayHello(HelloRequest) returns (HelloResponse) {}
}

message HelloRequest {
    string name = 1;
}

message HelloResponse {
    string message = 1;
}
`

var tmpDir = filepath.Join(os.TempDir(), "xxx__proto")

// create a test proto and return a teardown function.
func setup() (func() error, error) {
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return nil, err
	}
	err := ioutil.WriteFile(filepath.Join(tmpDir, "test.proto"), []byte(testProto), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return func() error {
		return os.RemoveAll(tmpDir)
	}, nil
}

// catch error and clearing test data.
func do(t *testing.T, setup func() (func() error, error), testFunc func(t *testing.T)) {
	teardown, err := setup()
	if err != nil {
		t.Fatalf("setup err:%s", err.Error())
		return
	}

	defer func() {
		if teardown != nil {
			if err := teardown(); err != nil {
				t.Fatalf("teardown err:%s", err.Error())
			}
		}
	}()

	testFunc(t)
}

func TestDescriptorSourceFromProtoFiles(t *testing.T) {
	do(t, setup, func(t *testing.T) {
		descriptor, err := DescriptorSourceFromProtoFiles([]string{tmpDir})
		if err != nil {
			t.Errorf("err:%s", err)
		}

		dsc, err := descriptor.FindSymbol("hello.HelloService.SayHello")
		if err != nil {
			t.Errorf("find symbol hello.HelloService.SayHello err:%s", err)
		}
		methodDesc := dsc.(*desc.MethodDescriptor)
		t.Logf("methodDesc: %v \n", methodDesc)
	})
}
