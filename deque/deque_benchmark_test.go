// Copyright 2026 zelr0x
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deque

import (
	stdlist "container/list"
	"testing"
	"time"

	. "github.com/zelr0x/chainyq/internal/benchutil"

	ed "github.com/edwingeng/deque/v2" // sync.pooled, segmented or similar
	gz "github.com/gammazero/deque"    // ring-buffer-based
	"github.com/zelr0x/chainyq/list"
)

const seed int64 = 31337

var Sink int
var BigStructSink BigStruct

type BigStruct struct {
	A, B, C, D, E, F, G, H int64
}

func BenchmarkPushBack(b *testing.B) {
	b.Run("chainyq.Deque", func(b *testing.B) {
		d := New[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("chainyq.Deque_Ensure", func(b *testing.B) {
		d := New[int]()
		d.EnsureBack(b.N)
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("edwingeng.Deque", func(b *testing.B) {
		d := ed.NewDeque[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("gammazero.Deque", func(b *testing.B) {
		var d gz.Deque[int]
		d.SetBaseCap(b.N)
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("gammazero.Deque_SetBaseCap", func(b *testing.B) {
		var d gz.Deque[int]
		d.SetBaseCap(b.N)
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("chainyq.list.List", func(b *testing.B) {
		d := list.New[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
	b.Run("container.list.List", func(b *testing.B) {
		d := stdlist.New()
		b.ResetTimer()
		for i := range b.N {
			d.PushBack(i)
		}
	})
}

func BenchmarkPushFront(b *testing.B) {
	b.Run("chainyq.Deque", func(b *testing.B) {
		d := New[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
	b.Run("chainyq.Deque_Ensure", func(b *testing.B) {
		d := New[int]()
		d.EnsureFront(b.N)
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
	b.Run("edwingeng.Deque", func(b *testing.B) {
		d := ed.NewDeque[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
	b.Run("gammazero.Deque", func(b *testing.B) {
		var d gz.Deque[int]
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
	b.Run("chainyq.list.List", func(b *testing.B) {
		d := list.New[int]()
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
	b.Run("container.list.List", func(b *testing.B) {
		d := stdlist.New()
		b.ResetTimer()
		for i := range b.N {
			d.PushFront(i)
		}
	})
}

func BenchmarkBlockBoundaryThrash(b *testing.B) {
	d := New[int]()
	di := d.Introspector()
	blockSize := di.BlockSize()
	for {
		_, off := di.BackIndex()
		if off == blockSize-1 {
			break
		}
		d.PushBack(0)
	}
	var sink int
	b.ResetTimer()

	for i := range b.N {
		d.PushBack(i)
		x, _ := d.PopBack()
		sink = x
	}
	Sink = sink
}

func BenchmarkChurn(b *testing.B) {
	b.Run("chainyq.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := New[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
		}
		Sink = sink
	})
	b.Run("chainyq.Deque_Pooled", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := NewPooled[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
		}
		Sink = sink
	})
	b.Run("chainyq.Deque_Ensure", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := New[int]()
		// b.N/3 instead of b.N/4 to have some slack for consistency
		d.EnsureFront(b.N / 3)
		d.EnsureBack(b.N / 3)
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
		}
		Sink = sink
	})
	b.Run("edwingeng.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := ed.NewDeque[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.TryPopBack()
				sink = x
			case 3:
				x, _ := d.TryPopFront()
				sink = x
			}
		}
		Sink = sink
	})
	b.Run("gammazero.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		var d gz.Deque[int]
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				if d.Len() != 0 {
					x := d.PopBack()
					sink = x
				}
			case 3:
				if d.Len() != 0 {
					x := d.PopFront()
					sink = x
				}
			}
		}
		Sink = sink
	})
	b.Run("chainyq.list.List", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := list.New[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
		}
		Sink = sink
	})
	b.Run("container.list.List", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		d := stdlist.New()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				if v := d.Back(); v != nil {
					x := d.Remove(v)
					sink = x.(int)
				}
			case 3:
				if v := d.Front(); v != nil {
					x := d.Remove(v)
					sink = x.(int)
				}
			}
		}
		Sink = sink
	})
}

func BenchmarkChurnWithClear(b *testing.B) {
	b.Run("chainyq.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/2)
		d := New[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("chainyq.Deque_Pooled", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/2)
		d := NewPooled[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("chainyq.Deque_PoolSmallPrealloc", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/10)
		d := WithCfg[int](DequeCfg{
			BlockSize: SuggestBlockSize[int](),
			FrontCap:  8,
			BackCap:   8,
			Pooled:    true,
		})
		d.PreallocPool(max(100, b.N/30)) // intentionally preallocate less than needed
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("edwingeng.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/10)
		d := ed.NewDeque[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.TryPopBack()
				sink = x
			case 3:
				x, _ := d.TryPopFront()
				sink = x
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("gammazero.Deque", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/10)
		var d gz.Deque[int]
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				if d.Len() != 0 {
					x := d.PopBack()
					sink = x
				}
			case 3:
				if d.Len() != 0 {
					x := d.PopFront()
					sink = x
				}
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("chainyq.list.List", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/10)
		d := list.New[int]()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
			if d.Len() > maxLen {
				d.Clear()
			}
		}
		Sink = sink
	})
	b.Run("container.list.List", func(b *testing.B) {
		a := RandomIntSlice(b, seed, 4)
		maxLen := max(1024, b.N/10)
		d := stdlist.New()
		var sink int
		b.ResetTimer()
		for i := range b.N {
			switch a[i] {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				if v := d.Back(); v != nil {
					x := d.Remove(v)
					sink = x.(int)
				}
			case 3:
				if v := d.Front(); v != nil {
					x := d.Remove(v)
					sink = x.(int)
				}
			}
			if d.Len() > maxLen {
				*d = *stdlist.New()
			}
		}
		Sink = sink
	})
}

func BenchmarkChurnBigStruct(b *testing.B) {
	b.Run("chainyq.Deque", func(b *testing.B) {
		ops := RandomIntSliceN(b, seed, b.N, 4)
		d := New[BigStruct]()
		var sink BigStruct
		b.ResetTimer()
		for i := range b.N {
			v := BigStruct{A: int64(i), B: int64(i)}
			switch ops[i] {
			case 0:
				d.PushBack(v)
			case 1:
				d.PushFront(v)
			case 2:
				x, _ := d.PopBack()
				sink = x
			case 3:
				x, _ := d.PopFront()
				sink = x
			}
		}
		BigStructSink = sink
	})
	b.Run("edwingeng.Deque", func(b *testing.B) {
		ops := RandomIntSliceN(b, seed, b.N, 4)
		d := ed.NewDeque[BigStruct]()
		var sink BigStruct
		b.ResetTimer()
		for i := range b.N {
			v := BigStruct{A: int64(i), B: int64(i)}
			switch ops[i] {
			case 0:
				d.PushBack(v)
			case 1:
				d.PushFront(v)
			case 2:
				x, _ := d.TryPopBack()
				sink = x
			case 3:
				x, _ := d.TryPopFront()
				sink = x
			}
		}
		BigStructSink = sink
	})
	b.Run("gammazero.Deque", func(b *testing.B) {
		ops := RandomIntSliceN(b, seed, b.N, 4)
		var d gz.Deque[BigStruct]
		var sink BigStruct
		b.ResetTimer()
		for i := range b.N {
			v := BigStruct{A: int64(i), B: int64(i)}
			switch ops[i] {
			case 0:
				d.PushBack(v)
			case 1:
				d.PushFront(v)
			case 2:
				if d.Len() != 0 {
					x := d.PopBack()
					sink = x
				}
			case 3:
				if d.Len() != 0 {
					x := d.PopFront()
					sink = x
				}
			}
		}
		BigStructSink = sink
	})
}

func BenchmarkRandomAccess(b *testing.B) {
	N := 1_000_000
	a := RandomIntSliceN(b, seed, N, N)

	b.Run("chainyq.Deque", func(b *testing.B) {
		d := New[int]()
		for i := range N {
			d.PushBack(i)
		}
		b.ResetTimer()
		var sink int
		j := 0
		for range b.N {
			target := a[j]
			x, _ := d.Get(target)
			sink = x
			j++
			if j == N {
				j = 0
			}
		}
		Sink = sink
	})

	b.Run("edwingeng.Deque", func(b *testing.B) {
		d := ed.NewDeque[int]()
		for i := range N {
			d.PushBack(i)
		}
		b.ResetTimer()
		var sink int
		j := 0
		for range b.N {
			target := a[j]
			x := d.Peek(target)
			sink = x
			j++
			if j == N {
				j = 0
			}
		}
		Sink = sink
	})

	b.Run("gammazero.Deque", func(b *testing.B) {
		var d gz.Deque[int]
		for i := range N {
			d.PushBack(i)
		}
		b.ResetTimer()
		var sink int
		j := 0
		for range b.N {
			target := a[j]
			x := d.At(target)
			sink = x
			j++
			if j == N {
				j = 0
			}
		}
		Sink = sink
	})
}

func BenchmarkBurstyQueue(b *testing.B) {
	type Event struct {
		ID   int
		Data [64]byte
	}
	const bursts = 100
	const writes = 100_000
	const reads = writes / 10
	totalOps := bursts*writes + bursts*reads
	b.Run("chainyq.Deque", func(b *testing.B) {
		d := New[Event]()
		var sink int
		b.ResetTimer()
		start := time.Now()
		for range bursts {
			for j := range writes {
				d.PushBack(Event{ID: j})
			}
			for range reads {
				v, ok := d.PopFront()
				if ok {
					sink = v.ID
				}
			}
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(totalOps), "ns/op")
		Sink = sink
	})

	b.Run("edwingeng.Deque", func(b *testing.B) {
		d := ed.NewDeque[Event]()
		var sink int
		b.ResetTimer()
		start := time.Now()
		for range bursts {
			for j := range writes {
				d.PushBack(Event{ID: j})
			}
			for range reads {
				v, ok := d.TryPopFront()
				if ok {
					sink = v.ID
				}
			}
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(totalOps), "ns/op")
		Sink = sink
	})

	b.Run("gammazero.Deque", func(b *testing.B) {
		var d gz.Deque[Event]
		var sink int
		b.ResetTimer()
		start := time.Now()
		for range bursts {
			for j := range writes {
				d.PushBack(Event{ID: j})
			}
			for range reads {
				if d.Len() > 0 {
					v := d.PopFront()
					sink = v.ID
				}
			}
		}
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(totalOps), "ns/op")
		Sink = sink
	})
}

func BenchmarkSlidingWindowMax(b *testing.B) {
	const N = 1_000_000
	const W = 1024
	const R = 10
	data := RandomIntSliceN(b, seed, N, N)

	b.Run("chainyq.Deque", func(b *testing.B) {
		d := New[int]()
		b.ResetTimer()
		var elapsedTotal time.Duration
		var sink int
		for range R {
			start := time.Now()
			for i := range N {
				for d.Len() > 0 {
					back, _ := d.Back()
					if data[back] >= data[i] {
						break
					}
					d.PopBack()
				}
				d.PushBack(i)
				front, _ := d.Front()
				if front <= i-W {
					d.PopFront()
				}
				if i >= W {
					x := data[front]
					sink += x
				}
			}
			elapsed := time.Since(start)
			elapsedTotal += elapsed
			Sink = sink
		}
		elapsed := elapsedTotal / R
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(N), "ns/op")
	})

	b.Run("edwingeng.Deque", func(b *testing.B) {
		d := ed.NewDeque[int]()
		b.ResetTimer()
		var elapsedTotal time.Duration
		var sink int
		for range R {
			start := time.Now()
			for i := range N {
				for d.Len() > 0 {
					back, _ := d.Back()
					if data[back] >= data[i] {
						break
					}
					d.PopBack()
				}
				d.PushBack(i)
				front, _ := d.Front()
				if front <= i-W {
					d.PopFront()
				}
				if i >= W {
					x := data[front]
					sink += x
				}
			}
			elapsed := time.Since(start)
			elapsedTotal += elapsed
			Sink = sink
		}
		elapsed := elapsedTotal / R
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(N), "ns/op")
	})

	b.Run("gammazero.Deque", func(b *testing.B) {
		var d gz.Deque[int]
		b.ResetTimer()
		var elapsedTotal time.Duration
		var sink int
		for range R {
			start := time.Now()
			for i := range N {
				for d.Len() > 0 {
					back := d.Back()
					if data[back] >= data[i] {
						break
					}
					d.PopBack()
				}
				d.PushBack(i)
				front := d.Front()
				if front <= i-W {
					d.PopFront()
				}
				if i >= W {
					x := data[front]
					sink += x
				}
			}
			elapsed := time.Since(start)
			elapsedTotal += elapsed
			Sink = sink
		}
		elapsed := elapsedTotal / R
		b.ReportMetric(float64(elapsed.Nanoseconds())/float64(N), "ns/op")
	})
}
