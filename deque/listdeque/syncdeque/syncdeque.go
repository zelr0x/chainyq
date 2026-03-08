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

// listdeque provides a list-based deque implementation.
package syncdeque

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zelr0x/chainyq/list"
)

// SyncListDeque is a list-based deque guarded by a sync.RWMutex.
// WARNING: SyncListDeque must not be copied after first use,
// because the embedded mutex would be copied and no longer
// protect the same state. Always use *SyncListDeque.
type SyncListDeque[T any] struct {
	l list.List[T]
	mu sync.RWMutex
}

// New creates a new deque.
func New[T any]() *SyncListDeque[T] {
	return &SyncListDeque[T]{l: list.NewValue[T]()}
}

// Len returns the length of a potentially nil deque. Returns 0 for nil.
func Len[T any](d *SyncListDeque[T]) int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.l.Len()
}

// String implements fmt.Stringer, used for %v and %+v.
func (d *SyncListDeque[T]) String() string {
	if d == nil {
		return "nil"
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.l.Len() == 0 {
		return "SyncListDeque[]"
	}
	var sb strings.Builder
	sb.WriteString("SyncListDeque[")
	it := d.l.Iter()
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
func (d *SyncListDeque[T]) GoString() string {
    if d == nil {
        return "nil"
    }
	var zero T
	d.mu.RLock()
	defer d.mu.RUnlock()
    if d.l.Len() == 0 {
        return fmt.Sprintf("SyncListDeque[%T]{}", zero)
    }
    var sb strings.Builder
    fmt.Fprintf(&sb, "SyncListDeque[%T]{", zero)
    it := d.l.Iter()
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
func (d *SyncListDeque[T]) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.l.Len()
}

// IsEmpty checks if the deque is empty.
func (d *SyncListDeque[T]) IsEmpty() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.l.IsEmpty()
}

// PushBack pushes the specified item to the back of the deque.
func (d *SyncListDeque[T]) PushBack(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.l.PushBack(v)
}

// PopBack removes and returns the item at the back end of the deque.
func (d *SyncListDeque[T]) PopBack() (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.l.PopBack()
}

// Back allows you to peek at the back of the deque.
func (d *SyncListDeque[T]) Back() (T, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.l.Back()
}

// PushFront pushes the specified item to the front of the deque.
func (d *SyncListDeque[T]) PushFront(v T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.l.PushFront(v)
}

// PopFront removes and returns the item at the front end of the deque.
func (d *SyncListDeque[T]) PopFront() (T, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.l.PopFront()
}

// Front allows you to peek at the front of the deque.
func (d *SyncListDeque[T]) Front() (T, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.l.Front()
}

// ToList converts this queue to the underlying list.
// Direct use of this list is not safe for concurrent access,
// and bypasses all synchronization provided by SyncListDeque.
//
// It is an escape hatch for when you need it. In many cases it indicates
// you should've used List from the beginning.
func (d *SyncListDeque[T]) ToListUnsafe() *list.List[T] {
	return &d.l
}
