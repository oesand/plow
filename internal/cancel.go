package internal

import (
	"context"
	"sync/atomic"
)

func CancellableDefer(fn func()) (do func(), include func(fn func()), cancel context.CancelFunc) {
	cd := &cancellableDefer{fn: fn}
	return cd.Defer, cd.Include, cd.Cancel
}

type cancellableDefer struct {
	done atomic.Bool
	fn   func()
	inc  []func()
}

func (c *cancellableDefer) Defer() {
	if c.done.Load() {
		return
	}
	c.done.Store(true)
	for _, fn := range c.inc {
		fn()
	}
	c.fn()
}

func (c *cancellableDefer) Include(fn func()) {
	if c.done.Load() {
		return
	}
	c.inc = append(c.inc, fn)
}

func (c *cancellableDefer) Cancel() {
	c.done.Store(true)
}
