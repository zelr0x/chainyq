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
	"unsafe"

	"github.com/zelr0x/chainyq/internal/numutil"
	"github.com/zelr0x/chainyq/internal/unsafeutil"
	"github.com/zelr0x/chainyq/seq"
)

const defInitCap = 32

// Stack is a classic slice-based stack.
type Stack[T any] struct {
	b []T
}

type Iter[T any] struct {
	s   []T
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
	return WrapValue(b)
}

// FromSlice creates a Stack by copying all the items from a given slice.
func FromSlice[T any](vals []T) *Stack[T] {
	b := make([]T, numutil.MaxInt(len(vals), 16))
	copy(b, vals)
	return Wrap(b)
}

// Wrap creates a Stack from a given slice by simply wrapping it.
// Mutations to one are seen in the other until the slice grows.
func Wrap[T any](slice []T) *Stack[T] {
	s := WrapValue(slice)
	return &s
}

// WrapValue creates a Stack from a given slice by simply wrapping it.
// Mutations to one are seen in the other until the slice grows.
func WrapValue[T any](slice []T) Stack[T] {
	return Stack[T]{
		b: slice,
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

// PushBack is an alias for [Push] added to conform to [chainyq.Stack].
func (s *Stack[T]) PushBack(v T) {
	s.Push(v)
}

// Pop removes and returns top item of the stack and true if the stack
// is not empty, zero value and false otherwise.
func (s *Stack[T]) Pop() (T, bool) {
	b := s.b
	n := len(b)
	if n == 0 {
		var zero T
		return zero, false
	}
	base := unsafe.SliceData(b)
	v := *unsafeutil.At(base, n-1)
	s.b = unsafe.Slice(base, n-1)
	return v, true
}

// PopBack is an alias for [Pop] added to conform to [chainyq.Stack].
func (s *Stack[T]) PopBack() (T, bool) {
	return s.Pop()
}

// Peek returns the top item of the stack and true if the stack is not empty,
// otherwise zero value and false. After that the item remains on the stack.
func (s *Stack[T]) Peek() (T, bool) {
	b := s.b
	n := len(b)
	if n == 0 {
		var zero T
		return zero, false
	}
	return *unsafeutil.At(unsafe.SliceData(b), n-1), true
}

// Back is an alias for [Peek] added to conform to [chainyq.Stack].
func (s *Stack[T]) Back() (T, bool) {
	return s.Peek()
}

func (s *Stack[T]) PeekPtr() (*T, bool) {
	b := s.b
	n := len(b)
	if n == 0 {
		return nil, false
	}
	return unsafeutil.At(unsafe.SliceData(b), n-1), true
}

// BackPtr is an alias for [PeekPtr] added to conform to [chainyq.Stack].
func (s *Stack[T]) BackPtr() (*T, bool) {
	return s.PeekPtr()
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

// ToSlice returns a slice of all items from top to bottom (logical order).
func (s *Stack[T]) ToSlice() []T {
	b := s.b
	n := len(b)
	if n == 0 {
		return []T{}
	}
	res := make([]T, 0, n)
	for i := n - 1; i >= 0; i-- {
		res = append(res, b[i])
	}
	return res
}

// Unwrap returns the copy of the underlying slice (physical order).
func (s *Stack[T]) UnwrapCopy() []T {
	b := s.b
	res := make([]T, len(b))
	copy(res, b)
	return res
}

// Returns the underlying slice directly.
func (s *Stack[T]) UnwrapUnsafe() []T {
	return s.b
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
