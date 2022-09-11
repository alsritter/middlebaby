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
	"fmt"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

// InitializeWatcher initialize a watcher to check for modification on all files
// in the given path to watch
func InitializeWatcher(pathToWatch ...string) (*watcher.Watcher, error) {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Create, watcher.Write)

	// add file or directory.
	for _, file := range pathToWatch {
		if err := w.AddRecursive(file); err != nil {
			return nil, fmt.Errorf("%v: error trying to watch change on %s directory", err, pathToWatch)
		}
	}

	return w, nil
}

// AttachWatcher start the watcher, if any error was produced while the starting process the application would crash
// you need to pass a function, this function is the function that will be executed when the watcher
// receive any event the type of defined on the InitializeWatcher function
func AttachWatcher(w *watcher.Watcher, fn func(event watcher.Event)) {
	go func() {
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()

	readEventsFromWatcher(w, fn)
}

func readEventsFromWatcher(w *watcher.Watcher, fn func(watcher.Event)) {
	go func() {
		for {
			select {
			case evt := <-w.Event:
				log.Println("Modified file:", evt.Path)
				fn(evt)
			case err := <-w.Error:
				log.Printf("Error checking file change: %+v", err)
			case <-w.Closed:
				return
			}
		}
	}()
}
