package ratukas

import (
	"time"
)

type Task struct {
	expiration int64
	callback   func()
}

func NewTask(expiration time.Time, callback func()) *Task {
	return &Task{
		expiration: expiration.UnixMilli(),
		callback:   callback,
	}
}
