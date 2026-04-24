package ratukas

import (
	"sync"
	"sync/atomic"
)

type Bucket struct {
	expiration atomic.Int64
	keys       []uint64
	mu         sync.Mutex
}

func NewBucket() *Bucket {
	b := &Bucket{
		keys: make([]uint64, 0, 16),
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

func (b *Bucket) Add(key uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.keys = append(b.keys, key)
}

func (b *Bucket) Flush() []uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	ks := b.keys
	b.keys = make([]uint64, 0, 16)
	b.expiration.Store(-1)

	return ks
}
