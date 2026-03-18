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

package syncdeque

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/zelr0x/chainyq/deque"
	. "github.com/zelr0x/chainyq/internal/testutil"
)

// go test -race -timeout 30s -run ^TestConcurrentAccess
func TestConcurrentAccess(t *testing.T) {
	d := New[int]()
	const N = 10000

	// PopFront + PushBack
	g1 := func() {
		for i := 0; i < N; i++ {
			if v, ok := d.PopFront(); ok {
				d.PushBack(v)
			} else {
				d.PushBack(i)
			}
		}
	}

	// PopBack + PushFront
	g2 := func() {
		for i := 0; i < N; i++ {
			if v, ok := d.PopBack(); ok {
				d.PushFront(v)
			} else {
				d.PushFront(i)
			}
		}
	}

	// Reads
	g3 := func() {
		for i := 0; i < N; i++ {
			_ = d.Len()
			_ = d.IsEmpty()
			d.Front()
			d.Back()
		}
	}

	done := make(chan struct{})
	go func() { g1(); close(done) }()
	go g2()
	go g3()

	select {
	case <-done:
		// success — one goroutine finished, others still running
	case <-time.After(5 * time.Second):
		t.Fatal("possible deadlock or hang")
	}
}

func TestSyncDequeConcurrentRandom(t *testing.T) {
	d := New[int]()
	intro := d.ToDequeUnsafe().Introspector()
	const N = 50_000

	done := make(chan struct{})
	go func() {
		for i := 0; i < N; i++ {
			switch i % 4 {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				d.PopBack()
			case 3:
				d.PopFront()
			}
		}
		close(done)
	}()

	go func() {
		for i := 0; i < N; i++ {
			_ = d.Len()
			_ = d.IsEmpty()
			_, _ = d.Front()
			_, _ = d.Back()
		}
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("possible deadlock or hang")
	}

	AssertSyncDequeInvariant(t, d, intro)
}

func TestSyncDequeClearRace(t *testing.T) {
	d := New[int]()
	intro := d.ToDequeUnsafe().Introspector()
	const N = 10_000

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := range N {
			d.PushBack(i)
		}
	}()

	go func() {
		defer wg.Done()
		for range N {
			d.PopFront()
		}
	}()

	go func() {
		defer wg.Done()
		for range N / 10 {
			d.Clear()
			d.ClearRelease()
		}
	}()

	wg.Wait()
	AssertSyncDequeInvariant(t, d, intro)
}

// ----- Helpers -----
func AssertSyncDequeInvariant[T any](
	t *testing.T,
	d *SyncDeque[T],
	intro deque.DequeIntrospector[T],
) {
	t.Helper()
	d.mu.RLock()
	defer d.mu.RUnlock()

	introMap := intro.Map()
	introLen := intro.Len()
	introMapLen := len(introMap)

	AssertTrue(t, introLen >= 0, "deque length is nonneg")
	AssertFalse(t, introLen != 0 && introMapLen == 0, fmt.Sprintf(
		"deque has zero blocks but non-zero length: %d", introLen))

	frontBlk, frontOff := intro.FrontIndex()
	backBlk, backOff := intro.BackIndex()
	blockSize := intro.BlockSize()

	if introLen == 0 {
		AssertEq(t, frontBlk, backBlk, "front and back blk must match for len=0")
		AssertEq(t, frontOff, backOff, "front and back offsets must match for len=0")
	}

	AssertTrue(t, frontOff >= 0, "front offset must be nonneg")
	AssertTrue(t, frontOff < blockSize, "front offset must fit block size")

	AssertTrue(t, backOff >= 0, "back offset must be nonneg")
	AssertTrue(t, backOff < blockSize, "back offset must fit block size")

	AssertTrue(t, frontBlk >= 0, "front blk must be nonneg")
	AssertTrue(t, backBlk >= 0, "back blk must be nonneg")

	AssertTrue(t, frontBlk < introMapLen, "front blk must fit map len")
	AssertTrue(t, backBlk < introMapLen, "back blk must fit map len")

	AssertTrue(t, frontBlk <= backBlk, "front blk <= back blk")

	AssertFalse(t, frontBlk == backBlk && frontOff > backOff,
		fmt.Sprintf("front offset > back offset within same block: front={blk: %v, off: %v}, back={%v, %v}",
			frontBlk, frontOff, backBlk, backOff))

	for i := frontBlk; i <= backBlk && i < introMapLen; i++ {
		AssertNotNil(t, introMap[i], fmt.Sprintf("block %d is nil but should exist", i))
	}
}
