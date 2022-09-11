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

package mbcontext

import (
	"context"
	"sync"
)

type Context struct {
	context.Context
	wg     *sync.WaitGroup
	cancel context.CancelFunc
}

type contextOptions struct {
}

// ContextOption sets options such as Isolate and Global Template to the NewContext
type ContextOption interface {
	apply(*contextOptions)
}

func NewContext(ctx context.Context) *Context {
	c, cancel := context.WithCancel(ctx)
	return &Context{wg: &sync.WaitGroup{}, Context: c, cancel: cancel}
}

func (c *Context) GetCancelFunc() context.CancelFunc {
	return c.cancel
}

func (c *Context) CancelFunc() {
	c.cancel()
}

func (c *Context) AddService(n int) {
	c.wg.Add(n)
}

func (c *Context) DoneService() {
	c.wg.Done()
}

func (c *Context) Wait() {
	c.wg.Wait()
}
