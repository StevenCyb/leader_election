package internal

import (
	"sync"
	"time"
)

// Timer struct.
type Timer struct {
	duration time.Duration
	trigger  func()
	timer    *time.Timer
	mu       sync.Mutex
	stopped  bool
}

// Set defines the timeout duration.
func (t *Timer) Set(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.duration = d
	t.stopped = false
}

// OnTrigger sets the function that will be called on timeout.
func (t *Timer) OnTrigger(triggerFunc func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.trigger = triggerFunc
}

// Reset restarts the timer with the previously set duration.
func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If the timer was stopped, allow it to restart
	t.stopped = false

	if t.timer != nil {
		t.timer.Stop()
	}

	t.start()
}

// Stop stops the timer
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timer != nil {
		t.timer.Stop()
		t.stopped = true
		t.timer = nil
	}
}

// start is an internal function that starts the timer.
func (t *Timer) start() {
	if t.stopped || t.duration == 0 || t.trigger == nil {
		return
	}

	t.timer = time.AfterFunc(t.duration, func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		if !t.stopped && t.trigger != nil {
			t.trigger()
		}
	})
}
