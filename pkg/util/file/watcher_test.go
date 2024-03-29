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

package file

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/radovskyb/watcher"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestAttachWatcher(t *testing.T) {
	var spy bool
	tests := []struct {
		name string
		w    *watcher.Watcher
		fn   func(watcher.Event)
	}{
		{"attach watcher and process", watcher.New(), func(watcher.Event) { spy = true }},
	}
	for _, tt := range tests {

		AttachWatcher(tt.w, tt.fn)
		tt.w.TriggerEvent(watcher.Create, nil)
		tt.w.Error <- errors.New("some error")
		tt.w.Close()
		time.Sleep(1 * time.Millisecond)
		if !spy {
			t.Error("can't read any events")
		}

	}
}

func TestInitializeWatcher(t *testing.T) {
	tests := []struct {
		name        string
		pathToWatch string
		wantWatcher bool
		wantErr     bool
	}{
		{"intialize valid watcher", "test/testdata.txt", true, false},
		{"invalid directory to watch", "<asdddee", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitializeWatcher(tt.pathToWatch)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitializeWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantWatcher && got == nil {
				t.Errorf("InitializeWatcher() got = %v, want a pointer watcher", got)
			}
		})
	}
}
