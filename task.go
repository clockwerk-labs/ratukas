package ratukas

import (
	"time"
)

type Task[T any] struct {
	expiration int64
	value      T
}

func NewTask[T any](expiration time.Time, value T) *Task[T] {
	return &Task[T]{
		expiration: expiration.UnixMilli(),
		value:      value,
	}
}
