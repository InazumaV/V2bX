package task

import (
	"sync"
	"time"
)

// Task is a task that runs periodically.
type Task struct {
	// Interval of the task being run
	Interval time.Duration
	// Execute is the task function
	Execute func() error

	access  sync.Mutex
	timer   *time.Timer
	running bool
}

func (t *Task) hasClosed() bool {
	t.access.Lock()
	defer t.access.Unlock()

	return !t.running
}

func (t *Task) checkedExecute(first bool) error {
	if t.hasClosed() {
		return nil
	}

	t.access.Lock()
	defer t.access.Unlock()
	if first {
		if err := t.Execute(); err != nil {
			t.running = false
			return err
		}
	}
	if !t.running {
		return nil
	}
	t.timer = time.AfterFunc(t.Interval, func() {
		t.checkedExecute(true)
	})

	return nil
}

// Start implements common.Runnable.
func (t *Task) Start(first bool) error {
	t.access.Lock()
	if t.running {
		t.access.Unlock()
		return nil
	}
	t.running = true
	t.access.Unlock()
	if err := t.checkedExecute(first); err != nil {
		t.access.Lock()
		t.running = false
		t.access.Unlock()
		return err
	}
	return nil
}

// Close implements common.Closable.
func (t *Task) Close() {
	t.access.Lock()
	defer t.access.Unlock()

	t.running = false
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}
