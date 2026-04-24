package ratukas

import (
	"sync/atomic"
	"time"
)

type Task struct {
	id         uint64
	expiration int64
	cancelled  atomic.Bool
}

func NewTask(id uint64, expiration time.Time) *Task {
	return &Task{
		id:         id,
		expiration: expiration.UnixMilli(),
	}
}

func (t *Task) Cancel() {
	t.cancelled.Store(true)
}

func (t *Task) IsCancelled() bool {
	return t.cancelled.Load()
}
