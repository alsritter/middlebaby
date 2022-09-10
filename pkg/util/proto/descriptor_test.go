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
