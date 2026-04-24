package ratukas

import (
	"sync"
	"sync/atomic"
)

type Bucket struct {
	expiration atomic.Int64
	tasks      []*Task
	mu         sync.Mutex
}

func NewBucket() *Bucket {
	b := &Bucket{
		tasks: make([]*Task, 0, 16),
	}

	b.expiration.Store(-1)

	return b
}

func (b *Bucket) Expiration() int64 {
	return b.expiration.Load()
}

func (b *Bucket) ExpireIn(exp int64) bool {
	return b.expiration.Swap(exp) != exp
}

func (b *Bucket) Add(task *Task) {
	b.mu.Lock()
	defer b.mu.Unlock()

	task.index = len(b.tasks)
	task.bucket = b
	b.tasks = append(b.tasks, task)
}

func (b *Bucket) Remove(task *Task) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if task.bucket != b || task.index < 0 || task.index >= len(b.tasks) {
		return
	}

	lastIdx := len(b.tasks) - 1

	if task.index < lastIdx {
		lastTask := b.tasks[lastIdx]
		b.tasks[task.index] = lastTask
		lastTask.index = task.index
	}

	b.tasks[lastIdx] = nil
	b.tasks = b.tasks[:lastIdx]

	task.index = -1
	task.bucket = nil
}

func (b *Bucket) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, task := range b.tasks {
		task.index = -1
		task.bucket = nil
	}

	b.tasks = make([]*Task, 0)
	b.expiration.Store(-1)
}
