package ratukas

import (
	"sync/atomic"
	"time"
)

type Task struct {
	id         uint64
	expiration int64
	cancelled  atomic.Bool
	callback   func()
}

func NewTask(id uint64, expiration time.Time, callback func()) *Task {
	return &Task{
		id:         id,
		expiration: expiration.UnixMilli(),
		callback:   callback,
	}
}

func (t *Task) ID() uint64 {
	return t.id
}

func (t *Task) Cancel() {
	t.cancelled.Store(true)
}

func (t *Task) IsCancelled() bool {
	return t.cancelled.Load()
}

func (t *Task) Execute() {
	t.callback()
}
