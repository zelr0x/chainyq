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

// Package stack defines an unbounded stack based on a slice.
package stack

import (
	"fmt"
	"slices"
	"strings"

	"github.com/zelr0x/chainyq/seq"
)

const defInitCap = 32

// Stack is a classic slice-based stack.
type Stack[T any] struct {
	b []T
}

type Iter[T any] struct {
	s []T
	cur int // index in the underlying slice just after the next item
}

// New creates a stack with default capacity of 32.
func New[T any]() *Stack[T] {
	r := NewValue[T](defInitCap)
	return &r
}

// NewValue creates a stack with the specified initial capacity
// and returns it as a value.
func NewValue[T any](initCap int) Stack[T] {
	b := make([]T, 0, initCap)
	return Stack[T]{
		b: b,
	}
}

// Len returns the length of a deque. Returns 0 for nil.
func Len[T any](d *Stack[T]) int {
	if d == nil {
		return 0
	}
	return d.Len()
}

// String implements fmt.Stringer, used for %v and %+v.
func (d *Stack[T]) String() string {
	if d == nil {
		return "nil"
	}
	if d.Len() == 0 {
		return "Stack[]"
	}
	var sb strings.Builder
	sb.Grow(7 + 2*d.Len())
	sb.WriteString("Stack[")
	it := d.Iter()
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
func (d *Stack[T]) GoString() string {
	if d == nil {
		return "nil"
	}
	var zero T
	if d.Len() == 0 {
		return fmt.Sprintf("Stack[%T]{}", zero)
	}
	var sb strings.Builder
	sb.Grow(32 + 3*d.Len())
	fmt.Fprintf(&sb, "Stack[%T]{", zero)
	it := d.Iter()
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

// Equals returns true if both stacks are nil or this stack
// is element-wise equal to the specified stack according to the specified
// equality function, false otherwise. Empty stack is not equal to nil stack.
func (d *Stack[T]) Equals(other *Stack[T], eq func(T, T) bool) bool {
	switch {
	case d == nil:
		return other == nil
	case other == nil:
		return false
	case d.Len() != other.Len():
		return false
	case d.Len() == 0:
		return true
	}
	for i, v := range d.b {
		otherV := other.b[i]
		if !eq(v, otherV) {
			return false
		}
	}
	return true
}

func (s *Stack[T]) Len() int {
	return len(s.b)
}

// IsEmpty checks if there are any items in the stack.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.b) == 0
}

// Push puts the specified item on top of the stack.
func (s *Stack[T]) Push(v T) {
	s.b = append(s.b, v)
}

// Pop removes and returns top item of the stack and true if the stack
// is not empty, zero value and false otherwise.
func (s *Stack[T]) Pop() (T, bool) {
	n := len(s.b)
	if n == 0 {
		var zero T
		return zero, false
	}
	v := s.b[n-1]
	s.b = s.b[:n-1]
	return v, true
}

// Peek returns the top item of the stack and true if the stack is not empty,
// otherwise zero value and false. After that the item remains on the stack.
func (s *Stack[T]) Peek() (T, bool) {
	n := len(s.b)
	if n == 0 {
		var zero T
		return zero, false
	}
	v := s.b[n-1]
	return v, true
}

func (s *Stack[T]) PeekPtr() (*T, bool) {
	n := len(s.b)
	if n == 0 {
		return nil, false
	}
	v := &s.b[n-1]
	return v, true
}

// Ensure increases the stack's capacity to fit cap n more items.
// It is guaranteed that the there will be no allocations before
// at least that many items are pushed onto the stack.
func (s *Stack[T]) Ensure(n int) {
	s.b = slices.Grow(s.b, n)
}

// Clear removes all items from the stack.
func (s *Stack[T]) Clear() {
	s.b = s.b[:0]
}

// Iter creates a new iterator for this stack. Traversal order is LIFO,
// i.e. reverse relative to the underlying slice.
func (s *Stack[T]) Iter() *Iter[T] {
	return &Iter[T]{s: s.b, cur: s.Len()}
}

// HasNext returns true if there are any items left to traverse,
// false otherwise.
func (it *Iter[T]) HasNext() bool {
	return it.cur > 0
}

// Next returns the next item and true if there are any items
// remaining on the stack, otherwise zero value and false.
// The vlaue is not removed from the underlying stack.
func (it *Iter[T]) Next() (T, bool) {
	if !it.HasNext() {
		var zero T
		return zero, false
	}
	it.cur--
	return it.s[it.cur], true
}

// NextPtr returns the pointer to the next item and true if there are any items
// remaining on the stack, otherwise nil and false.
// The vlaue is not removed from the underlying stack.
func (it *Iter[T]) NextPtr() (*T, bool) {
	if !it.HasNext() {
		return nil, false
	}
	it.cur--
	return &it.s[it.cur], true
}

// Seq creates a lazy sequence from this iterator.
func (it *Iter[T]) Seq() seq.Seq[T] {
	return seq.New(it.Next)
}

// PtrSeq creates a lazy sequence from this iterator.
func (it *Iter[T]) PtrSeq() seq.Seq[*T] {
	return seq.New(it.NextPtr)
}
