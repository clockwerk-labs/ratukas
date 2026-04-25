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
	Engine[T any] struct {
		wheel    *TimingWheel
		registry *Registry[T]
		logger   *slog.Logger
		expiry   <-chan *Bucket
		out      chan<- T
		pq       priorityQueue
		mu       sync.Mutex
	}

	bucketItem struct {
		bucket     *Bucket
		expiration int64
		index      int
	}

	priorityQueue []*bucketItem
)

func (pq *priorityQueue) Len() int {
	return len(*pq)
}

func (pq *priorityQueue) Less(i, j int) bool {
	return (*pq)[i].expiration < (*pq)[j].expiration
}

func (pq *priorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*bucketItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]

	return item
}

func NewEngine[T any](
	start time.Time,
	tick time.Duration,
	wheelSize int64,
	registry *Registry[T],
	logger *slog.Logger,
	out chan<- T,
) *Engine[T] {
	expiry := make(chan *Bucket)

	return &Engine[T]{
		wheel:    NewTimingWheel(start, tick, wheelSize, expiry),
		registry: registry,
		logger:   logger,
		expiry:   expiry,
		out:      out,
		pq:       make(priorityQueue, 0),
	}
}

func (e *Engine[T]) AddTask(key uint64, task *Task[T]) {
	e.registry.PutTask(key, task)
	e.wheel.Add(key, task.expiration)
}

func (e *Engine[T]) RemoveTask(key uint64) {
	e.registry.DeleteTask(key)
}

func (e *Engine[T]) Run(ctx context.Context) {
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

func (e *Engine[T]) resetTimer(t *time.Timer) {
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

func (e *Engine[T]) schedule(b *Bucket) {
	e.mu.Lock()
	defer e.mu.Unlock()

	heap.Push(&e.pq, &bucketItem{
		bucket:     b,
		expiration: b.Expiration(),
	})
}

func (e *Engine[T]) advance() {
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
				e.out <- task.value
			}
		}

		e.wheel.AdvanceTime(item.expiration)
	}
}
