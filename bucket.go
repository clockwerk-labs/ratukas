package ratukas

import (
	"sync"
	"sync/atomic"
)

type Bucket struct {
	expiration atomic.Int64
	tasks      []uint64
	mu         sync.Mutex
}

func NewBucket() *Bucket {
	b := &Bucket{
		tasks: make([]uint64, 0, 16),
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

	b.tasks = append(b.tasks, task.id)
}

func (b *Bucket) Flush() []uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	ts := b.tasks
	b.tasks = make([]uint64, 0, 16)
	b.expiration.Store(-1)

	return ts
}
