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
	"context"
	"testing"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
	"github.com/alsritter/middlebaby/pkg/util/synchronization"
)

func TestManager_GetMethod(t *testing.T) {
	var (
		ctx  = mbcontext.NewContext(context.Background())
		clog = logger.NewDefault("test")
	)

	pms, err := New(clog, &Config{
		ProtoImportPaths: []string{"temporary/alsritter/protobuf-examples"},
		SyncGitManger: &synchronization.Config{
			Enable:     true,
			StorageDir: "temporary",
			Repository: []*synchronization.Repository{{Address: "git@github.com:alsritter/protobuf-examples.git", Branch: "main"}},
		},
	})
	if err != nil {
		t.Error(err)
	}

	err = pms.Start(ctx)
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
