package ratukas

import (
	"time"
)

type Task struct {
	id         [16]byte
	expiration int64
	index      int
	bucket     *Bucket
}

func NewTask(id [16]byte, expiration time.Time) *Task {
	return &Task{
		id:         id,
		expiration: expiration.UnixMilli(),
		index:      -1,
	}
}
