package debounce

import (
	"github.com/bep/debounce"
	"time"
)

type Debouncer struct {
	handler    func(func())
	cancelFunc func()
}

func NewDebouncer(delay time.Duration) *Debouncer {
	hl, cancel := debounce.NewWithCancel(delay)
	return &Debouncer{
		handler:    hl,
		cancelFunc: cancel,
	}
}

func (d *Debouncer) Call(fn func(any2 any), args any) {
	d.handler(func() {
		fn(args)
	})
}

func (d *Debouncer) CancelAll() {
	d.cancelFunc()
}
