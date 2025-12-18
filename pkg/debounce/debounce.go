package debounce

import (
	"context"
	"sync"
	"time"

	"github.com/1pkg/gohalt"
)

type Debouncer struct {
	runner gohalt.Runner
	mu     sync.Mutex
	last   interface{}
}

func NewDebouncer(delay time.Duration) *Debouncer {
	throttler := gohalt.NewThrottlerTimed(20, delay, 0)
	return &Debouncer{
		runner: gohalt.NewRunnerAsync(context.Background(), throttler),
	}
}

func (d *Debouncer) Call(fn func(interface{}), arg interface{}) {
	d.mu.Lock()
	d.last = arg
	d.mu.Unlock()

	d.runner.Run(func(ctx context.Context) error {
		d.mu.Lock()
		arg := d.last
		d.mu.Unlock()
		fn(arg)
		return nil
	})
}
