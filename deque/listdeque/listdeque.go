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
package listdeque

import (
	"fmt"
	"strings"

	"github.com/zelr0x/chainyq/list"
)

// ListDeque is a list-based deque.
type ListDeque[T any] struct {
	l list.List[T]
}

// New creates a new deque.
func New[T any]() *ListDeque[T] {
	res := NewValue[T]()
	return &res
}

// NewValue returns an initialized deque. It can be useful for
// advanced use cases. Most users should prefer New().
func NewValue[T any]() ListDeque[T] {
	return ListDeque[T]{l: list.NewValue[T]()}
}

// Len returns the length of a potentially nil deque. Returns 0 for nil.
func Len[T any](d *ListDeque[T]) int {
	if d == nil {
		return 0
	}
	return d.l.Len()
}

// String implements fmt.Stringer, used for %v and %+v.
func (d *ListDeque[T]) String() string {
	if d == nil {
		return "nil"
	}
	if d.l.Len() == 0 {
		return "ListDeque[]"
	}
	var sb strings.Builder
	sb.Grow(11 + 2*d.l.Len())
	sb.WriteString("ListDeque[")
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
func (d *ListDeque[T]) GoString() string {
    if d == nil {
        return "nil"
    }
	var zero T
    if d.l.Len() == 0 {
        return fmt.Sprintf("ListDeque[%T]{}", zero)
    }
    var sb strings.Builder
	sb.Grow(32 + 3*d.l.Len())
    fmt.Fprintf(&sb, "ListDeque[%T]{", zero)
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
func (d *ListDeque[T]) Len() int {
	return d.l.Len()
}

// IsEmpty checks if the deque is empty.
func (d *ListDeque[T]) IsEmpty() bool {
	return d.l.IsEmpty()
}

// PushBack pushes the specified item to the back of the deque.
func (d *ListDeque[T]) PushBack(v T) {
	d.l.PushBack(v)
}

// PopBack removes and returns the item at the back end of the deque.
func (d *ListDeque[T]) PopBack() (T, bool) {
	return d.l.PopBack()
}

// Back allows you to peek at the back of the deque.
func (d *ListDeque[T]) Back() (T, bool) {
	return d.l.Back()
}

// PushFront pushes the specified item to the front of the deque.
func (d *ListDeque[T]) PushFront(v T) {
	d.l.PushFront(v)
}

// PopFront removes and returns the item at the front end of the deque.
func (d *ListDeque[T]) PopFront() (T, bool) {
	return d.l.PopFront()
}

// Front allows you to peek at the front of the deque.
func (d *ListDeque[T]) Front() (T, bool) {
	return d.l.Front()
}

// ToList converts this queue to the underlying list.
// It is an escape hatch for when you need it. In many cases it indicates
// you should've used List from the beginning.
func (d *ListDeque[T]) ToList() *list.List[T] {
	return &d.l
}
