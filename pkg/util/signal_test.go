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
	"os"
	"testing"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func TestRegisterExitHandlers(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	stop := RegisterExitHandlers(logger.NewDefault("test"), cancel)

	// do something...
	time.Sleep(2 * time.Second)
	sendInterruptSignal()

	<-stop

	clog := logger.NewDefault("test")
	clog.Info(nil, "server closed")
}

// ctrl + c
func sendInterruptSignal() error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(os.Interrupt)
}
