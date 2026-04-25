package ratukas

import (
	"sync/atomic"
	"time"
)

type TimingWheel struct {
	tick     int64
	size     int64
	interval int64
	buckets  []*Bucket
	expiry   chan<- *Bucket
	now      atomic.Int64
	overflow atomic.Pointer[TimingWheel]
}

func NewTimingWheel(start time.Time, tick time.Duration, size int64, expiry chan<- *Bucket) *TimingWheel {
	startMs, tickMs := start.UnixMilli(), tick.Milliseconds()

	tw := &TimingWheel{
		tick:     tickMs,
		size:     size,
		interval: tickMs * size,
		expiry:   expiry,
		buckets:  make([]*Bucket, size),
	}

	tw.now.Store(startMs - (startMs % tickMs))

	for i := range tw.buckets {
		tw.buckets[i] = NewBucket()
	}

	return tw
}

func (w *TimingWheel) Add(key uint64, expiration int64) bool {
	now := w.now.Load()

	if expiration < now+w.tick {
		return false
	}

	if expiration < now+w.interval {
		slot := expiration / w.tick
		bucket := w.buckets[slot%w.size]
		bucket.Add(key)
		if bucket.ExpireIn(slot * w.tick) {
			w.expiry <- bucket
		}

		return true
	}

	overflow := w.ascend()

	return overflow.Add(key, expiration)
}

func (w *TimingWheel) AdvanceTime(expiration int64) {
	for {
		now := w.now.Load()

		if expiration < now+w.tick {
			return
		}

		advanced := expiration - (expiration % w.tick)

		if w.now.CompareAndSwap(now, advanced) {
			if overflow := w.overflow.Load(); overflow != nil {
				overflow.AdvanceTime(advanced)
			}

			return
		}
	}
}

func (w *TimingWheel) ascend() *TimingWheel {
	if v := w.overflow.Load(); v != nil {
		return v
	}

	overflow := NewTimingWheel(
		time.UnixMilli(w.now.Load()),
		time.Duration(w.interval)*time.Millisecond,
		w.size,
		w.expiry,
	)

	if w.overflow.CompareAndSwap(nil, overflow) {
		return overflow
	}

	return w.overflow.Load()
}
