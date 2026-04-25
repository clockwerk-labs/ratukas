package ratukas

import (
	"fmt"
	"sync"
)

type (
	Registry[T any] struct {
		shards []*shard[T]
	}

	shard[T any] struct {
		tasks map[uint64]*Task[T]
		mu    sync.RWMutex
	}
)

func NewRegistry[T any](noOfShards int) *Registry[T] {
	shards := make([]*shard[T], noOfShards)

	for i := 0; i < noOfShards; i++ {
		shards[i] = &shard[T]{
			tasks: make(map[uint64]*Task[T]),
		}
	}

	return &Registry[T]{
		shards: shards,
	}
}

func (r *Registry[T]) GetTask(key uint64) (*Task[T], error) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[key]; !ok {
		return nil, fmt.Errorf("task %d not found", key)
	} else {
		return task, nil
	}
}

func (r *Registry[T]) PutTask(key uint64, task *Task[T]) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[key] = task
}

func (r *Registry[T]) DeleteTask(key uint64) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, key)
}
