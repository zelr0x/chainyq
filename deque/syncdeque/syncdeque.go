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

// package syncdeque provides [SyncDeque] - a synchronized deque based on a
// segmented array.
package syncdeque

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zelr0x/chainyq/deque"
)

// SyncDeque is a deque-based deque guarded by a sync.RWMutex.
// WARNING: SyncDeque must not be copied after first use,
// because the embedded mutex would be copied and no longer
// protect the same state. Always use *SyncDeque.
type SyncDeque[T any] struct {
	b  deque.Deque[T]
	mu sync.RWMutex
}

type SyncDequeCfg struct {
	deque.DequeCfg
}

// New creates a new deque.
func New[T any]() *SyncDeque[T] {
	d := deque.New[T]()
	return &SyncDeque[T]{b: *d}
}

func NewPooled[T any]() *SyncDeque[T] {
	d := deque.NewPooled[T]()
	return &SyncDeque[T]{b: *d}
}

func WithCfg[T any](cfg SyncDequeCfg) *SyncDeque[T] {
	d := deque.WithCfg[T](cfg.DequeCfg)
	return &SyncDeque[T]{b: *d}
}

// FromDeque creates a new SyncDeque by wrapping the specified Deque.
//
// The specified deque MUST NOT be used after this point. Using it
// concurrently with the SyncDeque will lead to internal state corruption.
//
// To recover the original deque, use the escape hatch [ToDequeUnsafe].
// After that, you may use the deque either by pointer or by value, but
// you MUST NOT use this SyncDeque again.
//
// If you need to jump back and forth between sync and unsync versions
// of the deque, you should use [ToDequeUnsafe] and [FromDeque] each time.
// Note that FromDeque copies the deque's bookkeeping structs (slice headers,
// allocator headers, indices, etc.) — this is cheap but not free, so it is
// optimized for prolonged use of the SyncDeque, not for frequent
// back-and-forth conversions.
func FromDeque[T any](d *deque.Deque[T]) *SyncDeque[T] {
	return &SyncDeque[T]{b: *d}
}

// Len returns the length of a potentially nil deque. Returns 0 for nil.
func Len[T any](d *SyncDeque[T]) int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.Len()
}

// String implements fmt.Stringer, used for %v and %+v.
func (d *SyncDeque[T]) String() string {
	if d == nil {
		return "nil"
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.b.Len() == 0 {
		return "SyncDeque[]"
	}
	var sb strings.Builder
	sb.Grow(15 + 2*d.b.Len())
	sb.WriteString("SyncDeque[")
	it := d.b.Iter()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		fmt.Fprintf(&sb, "%v", v)
		if !it.HasNext() {
			break
		}
		sb.WriteString(" ")
	}
	sb.WriteRune(']')
	return sb.String()
}

// GoString implements fmt.GoStringer, used for %#v.
func (d *SyncDeque[T]) GoString() string {
	if d == nil {
		return "nil"
	}
	var zero T
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.b.Len() == 0 {
		return fmt.Sprintf("SyncDeque[%T]{}", zero)
	}
	var sb strings.Builder
	sb.Grow(32 + 3*d.b.Len())
	fmt.Fprintf(&sb, "SyncDeque[%T]{", zero)
	it := d.b.Iter()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		fmt.Fprintf(&sb, "%#v", v)
		if !it.HasNext() {
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteRune('}')
	return sb.String()
}

// Len returns the number of items in the deque.
func (d *SyncDeque[T]) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.Len()
}

// IsEmpty checks if the deque is empty.
func (d *SyncDeque[T]) IsEmpty() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.IsEmpty()
}

// PushBack pushes the specified item to the back of the deque.
func (d *SyncDeque[T]) PushBack(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.PushBack(v)
}

// PopBack removes and returns the item at the back end of the deque.
func (d *SyncDeque[T]) PopBack() (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.b.PopBack()
}

// Back allows you to peek at the back of the deque.
func (d *SyncDeque[T]) Back() (T, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.Back()
}

// PushFront pushes the specified item to the front of the deque.
func (d *SyncDeque[T]) PushFront(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.PushFront(v)
}

// PopFront removes and returns the item at the front end of the deque.
func (d *SyncDeque[T]) PopFront() (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.b.PopFront()
}

// Front allows you to peek at the front of the deque.
func (d *SyncDeque[T]) Front() (T, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.Front()
}

func (d *SyncDeque[T]) Get(idx int) (T, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.b.Get(idx)
}

func (d *SyncDeque[T]) Set(idx int, v T) (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.b.Set(idx, v)
}

// EnsureFront grows the deque to allow at least the specified number of items
// to be pushed to the front before the next allocation occurs. Note that this
// does not change the empty space at the back, so if you need to push n items
// to the front and m items to the back before the next allocation, you should
// call EnsureFront(n) and by EnsureBack(m) together.
//
// If you're doing it right after deque creation, it is probably better
// to create unsync deque, ensure front/back and then convert it to sync deque.
//
// See also [deque.EnsureFront].
func (d *SyncDeque[T]) EnsureFront(items int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.EnsureFront(items)
}

// EnsureBack grows the deque to allow at least the specified number of items
// to be pushed to the back before the next allocation occurs. Note that this
// does not change the empty space at the front, so if you need to push n items
// to the front and m items to the back before the next allocation, you should
// call EnsureFront(n) and by EnsureBack(m) together.
//
// If you're doing it right after deque creation, it is probably better
// to create unsync deque, ensure front/back and then convert it to sync deque.
//
// See also [deque.EnsureBack].
func (d *SyncDeque[T]) EnsureBack(items int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.EnsureBack(items)
}

// ShrinkToFit shrinks the deque to the max of (used blocks, sun of initially
// configured front and back slack). If pooling is used, the memory is not
// reclaimed by the pool - the deque and the pool will not hold any pointers
// to the underlying memory beyond the amount required by the above formula.
func (d *SyncDeque[T]) ShrinkToFit() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.ShrinkToFit()
}

// Clear resets the deque without releasing any memory. This is useful
// when you want to reuse the deque without redundant reallocations.
// If the deque stores values and not pointers, make sure you don't have
// any live pointers to the items within the deque (create copies if you need)
// because the memory those pointers point to is up for reuse after this method
// is called. If you have many such live pointers, use [ClearRelease] instead.
func (d *SyncDeque[T]) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.Clear()
}

// ClearRelease discards all the allocated blocks and resets the deque
// to the initial state.
func (d *SyncDeque[T]) ClearRelease() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.b.ClearRelease()
}

// Todeque converts this queue to the underlying deque.
// Direct use of this deque is not safe for concurrent access,
// and bypasses all synchronization provided by SyncDeque.
//
// It is an escape hatch for when you need it. In many cases it indicates
// you should've used deque from the beginning.
func (d *SyncDeque[T]) ToDequeUnsafe() *deque.Deque[T] {
	return &d.b
}
