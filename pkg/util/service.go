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
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
)

// StartServiceAsync is used to start service async
func StartServiceAsync(ctx *mbcontext.Context, log logger.Logger,
	serveFn func() error, stopFn func() error) {
	if serveFn == nil {
		return
	}

	ctx.AddService(1)
	go func() {
		defer ctx.DoneService()

		log.Info(nil, "starting service")
		go func() {
			if err := serveFn(); err != nil {
				log.Error(nil, "error serving service: %s", err)
			}
			ctx.CancelFunc()
		}()

		<-ctx.Done()
		log.Info(nil, "stopping service")
		if stopFn() != nil {
			log.Info(nil, "stopping service gracefully")
			if err := stopFn(); err != nil {
				log.Warn(nil, "error occurred while stopping service: %s", err)
			}
		}
		log.Info(nil, "exiting service")
	}()
}
