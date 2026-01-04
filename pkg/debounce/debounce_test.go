package debounce

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebouncer_Call(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	var calls []any
	var mu sync.Mutex

	fn := func(arg any) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, arg)
	}

	// Call once
	debouncer.Call(fn, "arg1")

	// Wait for execution
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	assert.Len(t, calls, 1)
	assert.Equal(t, "arg1", calls[0])
	mu.Unlock()
}

func TestDebouncer_LastCallWins(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	var calls []any
	var mu sync.Mutex

	fn := func(arg any) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, arg)
	}

	// Call multiple times quickly
	debouncer.Call(fn, "arg1")
	debouncer.Call(fn, "arg2")
	debouncer.Call(fn, "arg3")

	// Wait for execution
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	// Depending on throttling, but at least the last should be called
	assert.Greater(t, len(calls), 0)
	assert.Equal(t, "arg3", calls[len(calls)-1]) // Last call should be arg3
	mu.Unlock()
}

func TestDebouncer_CancelAll(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	var calls []any
	var mu sync.Mutex

	fn := func(args any) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, args)
	}

	debouncer.Call(fn, "arg1")
	debouncer.Call(fn, "arg2")

	// Cancel all
	debouncer.CancelAll()

	// Wait
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	// Should not have called any
	assert.Len(t, calls, 0)
	mu.Unlock()
}
