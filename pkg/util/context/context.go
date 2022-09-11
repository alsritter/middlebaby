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

package context

import (
	"context"
	"sync"
	"time"
)

type Context struct {
	wg *sync.WaitGroup
}

func New(wg *sync.WaitGroup) context.Context { return &Context{wg: wg} }

// Deadline implements context.Context
func (*Context) Deadline() (deadline time.Time, ok bool) {
	panic("unimplemented")
}

// Done implements context.Context
func (*Context) Done() <-chan struct{} {
	panic("unimplemented")
}

// Err implements context.Context
func (*Context) Err() error {
	panic("unimplemented")
}

// Value implements context.Context
func (*Context) Value(key interface{}) interface{} {
	panic("unimplemented")
}
