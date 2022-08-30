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
