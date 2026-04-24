package ratukas

import (
	"container/heap"
	"context"
	"log/slog"
	"math"
	"sync"
	"time"
)

type (
	Engine struct {
		wheel    *TimingWheel
		registry *Registry
		logger   *slog.Logger
		expiry   <-chan *Bucket
		pq       PriorityQueue
		mu       sync.Mutex
	}

	BucketItem struct {
		bucket     *Bucket
		expiration int64
		index      int
	}

	PriorityQueue []*BucketItem
)

func (pq *PriorityQueue) Len() int {
	return len(*pq)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	return (*pq)[i].expiration < (*pq)[j].expiration
}

func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*BucketItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]

	return item
}

func NewEngine(wheel *TimingWheel, registry *Registry, logger *slog.Logger, expiry <-chan *Bucket) *Engine {
	return &Engine{
		wheel:    wheel,
		registry: registry,
		logger:   logger,
		pq:       make(PriorityQueue, 0),
		expiry:   expiry,
	}
}

func (e *Engine) AddTask(key uint64, task *Task) {
	e.registry.PutTask(key, task)
	e.wheel.Add(key, task)
}

func (e *Engine) RemoveTask(key uint64) {
	e.registry.DeleteTask(key)
}

func (e *Engine) Run(ctx context.Context) {
	timer := time.NewTimer(math.MaxInt)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case b := <-e.expiry:
			e.schedule(b)
			e.resetTimer(timer)
		case <-timer.C:
			e.advance()
			e.resetTimer(timer)
		}
	}
}

func (e *Engine) resetTimer(t *time.Timer) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.pq) == 0 {
		return
	}

	diff := e.pq[0].expiration - time.Now().UnixMilli()
	if diff <= 0 {
		diff = 0
	}

	t.Reset(time.Duration(diff) * time.Millisecond)
}

func (e *Engine) schedule(b *Bucket) {
	e.mu.Lock()
	defer e.mu.Unlock()

	heap.Push(&e.pq, &BucketItem{
		bucket:     b,
		expiration: b.Expiration(),
	})
}

func (e *Engine) advance() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UnixMilli()

	for e.pq.Len() > 0 {
		item := e.pq[0]

		if item.expiration > now {
			break
		}

		heap.Pop(&e.pq)

		for _, key := range item.bucket.Flush() {
			if task, err := e.registry.GetTask(key); err != nil {
				e.logger.Error("Failed to get task from registry", "key", key, "err", err)
			} else {
				task.callback()
			}
		}

		e.wheel.AdvanceTime(item.expiration)
	}
}
