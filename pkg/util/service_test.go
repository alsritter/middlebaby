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

package util

import (
	"context"
	"testing"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
)

func TestStartServiceAsync(t *testing.T) {
	ctx := mbcontext.NewContext(context.Background())
	clog := logger.NewDefault("test")
	StartServiceAsync(ctx, clog, func() error {
		// Here is the initialization project
		clog.Info(nil, "TestServer Starting...")
		return nil
	}, func() error {
		// Call if cancel is closed
		clog.Info(nil, "TestServer Closed...")
		return nil
	})

	time.Sleep(time.Second * 1)

	// close.
	ctx.CancelFunc()
	time.Sleep(time.Second * 2)
}

// Reference:
// https://stackoverflow.com/questions/66833138/wait-for-context-done-channel-for-cancellation-while-working-on-long-run-operati
func TestContext(t *testing.T) {
	var (
		workTimeCost  = 2 * time.Second
		cancelTimeout = 1 * time.Second
	)

	ctx, cancel := context.WithCancel(context.Background())

	var (
		data   int
		readCh = make(chan struct{})
	)

	go func() {
		defer close(readCh)
		t.Log("blocked to read data")
		// fake long i/o operations
		time.Sleep(workTimeCost)
		data = 10
		t.Log("done read data")
	}()

	// fake cancel is called from the other routine (it's actually not caused by timeout)
	time.AfterFunc(cancelTimeout, cancel)

	select {
	case <-ctx.Done():
		t.Log("cancelled")
		return
	case <-readCh:
		break
	}

	t.Log("got final data", data)
}
