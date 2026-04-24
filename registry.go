package ratukas

import (
	"fmt"
	"sync"
)

type (
	Registry struct {
		shards []*shard
	}

	shard struct {
		tasks map[uint64]*Task
		mu    sync.RWMutex
	}
)

func NewRegistry(noOfShards int) *Registry {
	shards := make([]*shard, noOfShards)

	for i := 0; i < noOfShards; i++ {
		shards[i] = &shard{
			tasks: make(map[uint64]*Task),
		}
	}

	return &Registry{
		shards: shards,
	}
}

func (r *Registry) GetTask(key uint64) (*Task, error) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[key]; !ok {
		return nil, fmt.Errorf("task %d not found", key)
	} else {
		return task, nil
	}
}

func (r *Registry) PutTask(key uint64, task *Task) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[key] = task
}

func (r *Registry) DeleteTask(key uint64) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, key)
}
