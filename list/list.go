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

package list

import (
	"fmt"
	"strings"
)

// TODO: accept config that defines allocation sizes
var sliceResultCap int = 32

// List is a doubly-linked list. Not thread-safe.
type List[T any] struct {
	head *node[T]
	tail *node[T]
	len  int
}

// Iter is an iterator for a List. Not thread-safe.
type Iter[T any] struct {
	l   *List[T]
	cur *node[T]
}

type node[T any] struct {
	prev *node[T]
	next *node[T]
	val  T
}

// New creates a new List.
func New[T any]() *List[T] {
	head := new(node[T])
	tail := new(node[T])
	head.next = tail
	tail.prev = head
	return &List[T]{head: head, tail: tail}
}

// FromSlice creates a new list from the specified slice.
func FromSlice[T any](vals []T) *List[T] {
	list := New[T]()
	for _, v := range vals {
		list.PushBack(v)
	}
	return list
}

// Len returns the length of a potentially nil list. Returns 0 for nil.
func Len[T any](list *List[T]) int {
	if list == nil {
		return 0
	}
	return list.len
}

// Append adds an item to a potentially nil list, returning the list back.
// If the specified list is nil, creates an empty list and appends to it.
func Append[T any](list *List[T], v T) *List[T] {
	if list == nil {
		list = New[T]()
	}
	list.Add(v)
	return list
}

// Len returns the number of items in the list.
func (l *List[T]) Len() int {
	return l.len
}

// IsEmpty check if the list is empty.
func (l *List[T]) IsEmpty() bool {
	return l.len == 0
}

// String is a Stringer implementation.
func (l *List[T]) String() string {
	if l == nil {
		return "nil"
	}
	if l.len == 0 {
		return "List[]"
	}
	var sb strings.Builder
	sb.WriteString("List[")
	it := l.Iter()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		fmt.Fprintf(&sb, "%v", v)
		if !it.HasNext() {
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteRune(']')
	return sb.String()
}

// Equals returns true if this list is element-wise equal to the specified list
// according to the specified equality function, false otherwise.
func (l *List[T]) Equals(other *List[T], eq func(T, T) bool) bool {
	if l.Len() != other.Len() {
		return false
	}
	if l.Len() == 0 {
		return true
	}
	itA := l.Iter()
	itB := other.Iter()
	var a, b T
	var okA, okB bool
	for {
		a, okA = itA.Next()
		b, okB = itB.Next()
		if !okA || !okB {
			return !okA && !okB
		}
		if !eq(a, b) {
			return false
		}
	}
}

// Front returns the first item and true if possible, zero value
// and false otherwise.
func (l *List[T]) Front() (T, bool) {
	if l.len == 0 {
		var zero T
		return zero, false
	}
	return l.head.next.val, true
}

// FrontPtr returns a pointer to the first item and true if possible,
// nil and false otherwise.
func (l *List[T]) FrontPtr() (*T, bool) {
	if l.len == 0 {
		return nil, false
	}
	return &l.head.next.val, true
}

// Back returns the last item and true if possible, zero value
// and false otherwise.
func (l *List[T]) Back() (T, bool) {
	if l.len == 0 {
		var zero T
		return zero, false
	}
	return l.tail.prev.val, true
}

// BackPtr returns a pointer to the last item and true if possible,
// nil and false otherwise.
func (l *List[T]) BackPtr() (*T, bool) {
	if l.len == 0 {
		return nil, false
	}
	return &l.tail.prev.val, true
}

// PushFront pushes an item to the front of the list.
func (l *List[T]) PushFront(val T) {
	l.insertAfter(l.head, l.newNode(val))
}

// PushBack pushes an item to the back of the list.
func (l *List[T]) PushBack(val T) {
	l.insertBefore(l.tail, l.newNode(val))
}

// Adds pushes an item to the back of the list, returning the list itself.
// This is the same as [PushBack], but can be chained.
func (l *List[T]) Add(val T) *List[T] {
	l.PushBack(val)
	return l
}

// Pop removes and returns the first item and true if possible,
// zero value and false otherwise.
func (l *List[T]) PopFront() (T, bool) {
	return l.remove(l.head.next)
}

// PopBack removes and returns the last item and true if possible,
// zero value and false otherwise.
func (l *List[T]) PopBack() (T, bool) {
	return l.remove(l.tail.prev)
}

// Insert adds an item at the specified index if possible, returning
// true on success or false otherwise.
func (l *List[T]) Insert(idx int, val T) bool {
	if idx < 0 || idx > l.len {
		return false
	}
	newNode := l.newNode(val)
	if idx == l.len {
		l.insertBefore(l.tail, newNode)
		return true
	}
	n := l.nodeAtUnchecked(idx)
	l.insertBefore(n, newNode)
	return true
}

// Remove removes the item with a given index from the list and returns
// true if found, false otherwise.
func (l *List[T]) Remove(idx int) bool {
	n := l.nodeAt(idx)
	if n == nil {
		return false
	}
	l.remove(n)
	return true
}

// IndexOf returns the index of the first occurrence of a target
// in the list, starting search at the front (head) of the list.
// eq is a function that checks if two items are equal.
func (l *List[T]) IndexOf(target T, eq func(T, T) bool) int {
	if Len(l) < 1 {
		return -1
	}
	it := l.Iter()
	i := 0
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		if eq(v, target) {
			return i
		}
		i++
	}
	return -1
}

// LastIndexOf returns the index of the first occurrence of a target
// in the list, starting search at the back (tail) of the list.
// eq is a function that checks if two items are equal.
func (l *List[T]) LastIndexOf(target T, eq func(T, T) bool) int {
	if Len(l) < 1 {
		return -1
	}
	it := l.Iter().ResetBack()
	i := l.len - 1
	for v, ok := it.Prev(); ok; v, ok = it.Prev() {
		if eq(v, target) {
			return i
		}
		i--
	}
	return -1
}

// Contains checks if a specified item exists in the list.
// eq is a function that checks if two items are equal.
func (l *List[T]) Contains(target T, eq func(T, T) bool) bool {
	return l.IndexOf(target, eq) != -1
}

// Get retrieves an item by index, returning an item and true if found,
// zero value and false otherwise.
func (l *List[T]) Get(idx int) (T, bool) {
	n := l.nodeAt(idx)
	if n == nil {
		var zero T
		return zero, false
	}
	return n.val, true
}

// ForEach calls the specified function for each item.
func (l *List[T]) ForEach(f func(T) bool) {
	l.Iter().ForEach(f)
}

// ForEachPtr calls the specified function for each item.
func (l *List[T]) ForEachPtr(f func(*T) bool) {
	l.Iter().ForEachPtr(f)
}

// RemoveIf removes all items that match the specified predicate.
func (l *List[T]) RemoveIf(pred func(T) bool) int {
	it := l.Iter()
	count := 0
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		if !pred(v) {
			continue
		}
		it.Remove()
		count++
	}
	return count
}

// Concat appends all elements of 'other' to this list in O(1) by splicing
// nodes. After this call, 'other' is emptied and must not be used.
func (l *List[T]) Concat(other *List[T]) {
	if other.IsEmpty() {
		return
	}
	if l.IsEmpty() {
		l.head, l.tail = other.head, other.tail
	} else {
		currentLast := l.tail.prev
		otherFirst := other.head.next

		currentLast.next = otherFirst
		otherFirst.prev = currentLast
		l.tail = other.tail
	}
	l.len += other.len
	other.Clear()
}

// Clear clears the list in a fast manner, linking head to tail and vice versa.
// The next GC should free all the items that don't have live references.
func (l *List[T]) Clear() {
	l.head.next = l.tail
	l.tail.prev = l.head
	l.len = 0
}

// Iter creates an iterator over the list, starting at its head. Operations
// such as [Next], [Skip], and some others advance the iterator by one or
// more positions. All operations work on the remaining portion of the list.
// When an operation can mutate the underlying list, this is explicitly
// stated in the doc comment of the corresponding method, except for the
// -Ptr methods - those can always mutate the values stored in the underlying
// list because they work on pointers.
func (l *List[T]) Iter() *Iter[T] {
	return l.newIter(l.head)
}

// Slice creates a [start, end) slice of the list if possible.
// On success, the length of the resulting slice is min(Len(), end-start).
// Returns an empty slice if the range is invalid.
func (l *List[T]) Slice(start, end int) []T {
	if start < 0 || end < 1 {
		return []T{}
	}
	// TODO: make it skip from the end if start is closer to the end.
	n := end - start
	if n < 1 {
		return []T{}
	}
	return l.Iter().Skip(start).TakeSlice(n)
}

// SlicePtr creates a [start, end) slice of the list if possible.
// On success, the length of the resulting slice is min(Len(), end-start).
// Returns an empty slice if the range is invalid.
func (l *List[T]) PtrSlice(start, end int) []*T {
	if start < 0 || end < 1 {
		return []*T{}
	}
	// TODO: make it skip from the end if start is closer to the end.
	n := end - start
	if n < 1 {
		return []*T{}
	}
	return l.Iter().Skip(start).TakePtrSlice(n)
}

// ToSlice creates a slice with all the items of the list.
func (l *List[T]) ToSlice() []T {
	return l.Slice(0, l.Len())
}

// ToPtrSlice creates a slice with all the items of the list.
func (l *List[T]) ToPtrSlice() []*T {
	return l.PtrSlice(0, l.Len())
}

// ToChan creates a read channel with all the items of the list.
func (l *List[T]) ToChan(size int) <-chan T {
	return l.Iter().ToChan(size)
}

// ToPtrChan creates a read channel with all the items of the list.
func (l *List[T]) ToPtrChan(size int) <-chan *T {
	return l.Iter().ToPtrChan(size)
}

func (l *List[T]) newNode(val T) *node[T] {
	return &node[T]{val: val}
}

func (l *List[T]) nodeAt(idx int) *node[T] {
	if idx < 0 || idx >= l.len {
		return nil
	}
	return l.nodeAtUnchecked(idx)
}

func (l *List[T]) nodeAtUnchecked(idx int) *node[T] {
	var n *node[T]
	if 2*idx >= l.len {
		n = l.tail.prev
		for i := l.len - 1; i > idx; i-- {
			n = n.prev
		}
	} else {
		n = l.head.next
		for i := 0; i < idx; i++ {
			n = n.next
		}
	}
	return n
}

func (l *List[T]) insertAfter(at *node[T], n *node[T]) {
	l.insertBefore(at.next, n)
}

func (l *List[T]) insertBefore(at *node[T], n *node[T]) {
	n.prev = at.prev
	n.next = at
	at.prev.next = n
	at.prev = n
	l.len++
}

func (l *List[T]) remove(n *node[T]) (T, bool) {
	if n == l.head || n == l.tail {
		return n.val, false
	}
	l.removeUnchecked(n)
	return n.val, true
}

func (l *List[T]) removeUnchecked(n *node[T]) {
	n.prev.next = n.next
	n.next.prev = n.prev
	l.len--
}

func (l *List[T]) newIter(start *node[T]) *Iter[T] {
	return &Iter[T]{l: l, cur: start}
}

// Clone creates a new iterator to the same underlying list starting
// at the current position of this iterator.
func (it *Iter[T]) Clone() *Iter[T] {
	return it.l.newIter(it.cur)
}

// Reset sets this iterator to point to the head of the underlying list
// and returns the iterator itself.
func (it *Iter[T]) Reset() *Iter[T] {
	it.cur = it.l.head
	return it
}

// ResetBack sets this iterator to point to the back (tail) of the underlying
// list and returns the iterator itself.
func (it *Iter[T]) ResetBack() *Iter[T] {
	it.cur = it.l.tail
	return it
}

// Current returns the current item and true if [Next] was called at least once
// on a non-empty list, zero value and false otherwise.
//
// This is intended for special introspection/debugging purposes, not as the main
// traversal method - for the main traversal use [Next].
func (it *Iter[T]) Current() (T, bool) {
	cur := it.cur
	if cur == it.l.head || cur == it.l.tail {
		var zero T
		return zero, false
	}
	return cur.val, true
}

// Current returns the current item and true if [Next] was called at least once
// on a non-empty list, zero value and false otherwise.
//
// This is intended for special introspection/debugging purposes, not as the main
// traversal method - for the main traversal use [Next].
func (it *Iter[T]) CurrentPtr() (*T, bool) {
	cur := it.cur
	if cur == it.l.head || cur == it.l.tail {
		return nil, false
	}
	return &cur.val, true
}

// HasNext reports whether there is a next item.
//
// You only need to call this method explicitly if you want to check for
// remaining items yourself. Most iterator methods call it internally.
func (it *Iter[T]) HasNext() bool {
	// INVARIANT: it.cur should never become it.l.tail.
	// is kind of a lie. It enforces good method design though.
	next := it.cur.next
	return next != nil && next != it.l.tail
}

// HasPrev reports whether there is a previous item.
func (it *Iter[T]) HasPrev() bool {
	prev := it.cur.prev
	return prev != nil && prev != it.l.head
}

// Next returns the next item and true if the item is available,
// zero value and false otherwise. The iterator advances by one position.
func (it *Iter[T]) Next() (T, bool) {
	v, ok := it.Peek()
	if ok {
		it.advance() // INVARIANT: it.cur should never become it.l.tail
	}
	return v, ok
}

// NextPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise. The iterator advances by one position.
func (it *Iter[T]) NextPtr() (*T, bool) {
	// it is a contract that it.cur is never it.l.tail
	v, ok := it.PeekPtr()
	if ok {
		it.advance() // INVARIANT: it.cur should never become it.l.tail
	}
	return v, ok
}

// Prev returns the previous item and true if the item is available,
// zero value and false otherwise.
func (it *Iter[T]) Prev() (T, bool) {
	v, ok := it.PeekBack()
	if ok {
		it.stepBack()
	}
	return v, ok
}

// PrevPtr returns the previous item and true if the item is available,
// zero value and false otherwise.
func (it *Iter[T]) PrevPtr() (*T, bool) {
	v, ok := it.PeekBackPtr()
	if ok {
		it.stepBack()
	}
	return v, ok
}

// Peek returns the next item and true if it is available,
// zero value and false otherwise.
func (it *Iter[T]) Peek() (T, bool) {
	if !it.HasNext() {
		var zero T
		return zero, false
	}
	return it.cur.next.val, true
}

// PeekPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise.
func (it *Iter[T]) PeekPtr() (*T, bool) {
	if !it.HasNext() {
		return nil, false
	}
	return &it.cur.next.val, true
}

// PeekBack returns the previous item and true if it is available,
// zero value and false otherwise.
func (it *Iter[T]) PeekBack() (T, bool) {
	if !it.HasPrev() {
		var zero T
		return zero, false
	}
	return it.cur.prev.val, true
}

// PeekBackPtr returns a pointer to the previous item and true
// if it is available, nil and false otherwise.
func (it *Iter[T]) PeekBackPtr() (*T, bool) {
	if !it.HasPrev() {
		return nil, false
	}
	return &it.cur.prev.val, true
}

// Remove removes the current item from the underlying list and returns
// it and true if the operation succeeded, otherwise zero value and false.
// The iterator is then repositioned at the predeccessor of the removed item.
// A subsequent call to Next() will return the successor to the removed item.
// Returns false if there's nothing to remove i.e. the iterator points to either
// the head or tail of the underlying list. Modifies the underlying list.
func (it *Iter[T]) Remove() (T, bool) {
	if it.cur == it.l.head || it.cur == it.l.tail {
		var zero T
		return zero, false
	}
	rem := it.cur
	it.stepBack()
	return it.l.remove(rem)
}

// InsertBefore inserts the given item before the current item. If called on an
// iterator that is currently at the head of the list, false is returned.
// Modifies the underlying list.
func (it *Iter[T]) InsertBefore(val T) bool {
	if it.cur == it.l.head {
		return false
	}
	newNode := it.l.newNode(val)
	it.l.insertBefore(it.cur, newNode)
	return true
}

// InsertAfter inserts the given item after the current item. If called on an
// iterator that is currently at the tail of the list, false is returned.
func (it *Iter[T]) InsertAfter(val T) bool {
	if it.cur == it.l.tail {
		return false
	}
	newNode := it.l.newNode(val)
	it.l.insertAfter(it.cur, newNode)
	return true
}

// ForEach calls the specified function for each remaining item.
func (it *Iter[T]) ForEach(f func(T) bool) {
	for {
		v, ok := it.Next()
		if !ok || !f(v) {
			break
		}
	}
}

// ForEachPtr calls the specified function for each remaining item.
func (it *Iter[T]) ForEachPtr(f func(val *T) bool) {
	for {
		v, ok := it.NextPtr()
		if !ok || !f(v) {
			break
		}
	}
}

// ToChan consumes the iterator and returns a channel of specified size
// to which all next items starting from the current position will be written.
func (it *Iter[T]) ToChan(size int) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		for v, ok := it.Next(); ok; v, ok = it.Next() {
			ch <- v
		}
	}()
	return ch
}

// ToPtrChan consumes the iterator and returns a read channel of specified size
// to which all next items starting from the current position will be written.
func (it *Iter[T]) ToPtrChan(size int) <-chan *T {
	ch := make(chan *T, size)
	go func() {
		defer close(ch)
		for v, ok := it.NextPtr(); ok; v, ok = it.NextPtr() {
			ch <- v
		}
	}()
	return ch
}

// TakeSlice collects up to n items to a slice, starting from the current
// position. Returns an empty slice if the list is empty.
func (it *Iter[T]) TakeSlice(n int) []T {
	if !it.HasNext() {
		return []T{}
	}
	res := make([]T, 0, min(it.l.len, n))
	i := 0
	for v, ok := it.Next(); ok && i < n; v, ok = it.Next() {
		res = append(res, v)
		i++
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakePtrSlice collects up to n items to a slice, starting from the current
// position. Returns an empty slice if the list is empty.
func (it *Iter[T]) TakePtrSlice(n int) []*T {
	if !it.HasNext() {
		return []*T{}
	}
	res := make([]*T, 0, min(it.l.len, n))
	i := 0
	for v, ok := it.NextPtr(); ok && i < n; v, ok = it.NextPtr() {
		res = append(res, v)
		i++
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// TakeWhile collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns empty slice if the list
// is empty.
func (it *Iter[T]) TakeWhile(pred func(T) bool) []T {
	if !it.HasNext() {
		return []T{}
	}
	res := make([]T, 0, min(it.l.len, sliceResultCap))
	for v, ok := it.Next(); ok && pred(v); v, ok = it.Next() {
		res = append(res, v)
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakeWhilePtr collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns nil slice as empty.
// eq is a function that checks if two items are equal.
func (it *Iter[T]) TakeWhilePtr(pred func(*T) bool) []*T {
	if !it.HasNext() {
		return []*T{}
	}
	res := make([]*T, 0, min(it.l.len, sliceResultCap))
	for v, ok := it.NextPtr(); ok && pred(v); v, ok = it.NextPtr() {
		res = append(res, v)
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// Skip advances the iterator by a specified number of positions (or fewer
// if fewer items remain) and returns the iterator itself.
func (it *Iter[T]) Skip(n int) *Iter[T] {
	for i := 0; i < n && it.HasNext(); i++ {
		it.advance() // INVARIANT: it.cur should never become it.l.tail
	}
	return it
}

// SkipBack steps back the iterator by a specified number of positions (or fewer
// if fewer items are before the current position) and returns the iterator
// itself.
func (it *Iter[T]) SkipBack(n int) *Iter[T] {
	for i := 0; i < n && it.HasPrev(); i++ {
		it.stepBack()
	}
	return it
}

// SkipWhile advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself stopped
// at the last element that matched the predicate - the first item that fails
// the predicate will be returned by the first call to [Next].
func (it *Iter[T]) SkipWhile(pred func(T) bool) *Iter[T] {
	for v, _ := it.Peek(); it.HasNext() && pred(v); v, _ = it.Peek() {
		it.advance() // INVARIANT: it.cur should never become it.l.tail
	}
	return it
}

// SkipWhilePtr advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself stopped
// at the last element that matched the predicate - the first item that fails
// the predicate will be returned by the first call to [Next].
func (it *Iter[T]) SkipWhilePtr(pred func(*T) bool) *Iter[T] {
	for v, _ := it.PeekPtr(); it.HasNext() && pred(v); v, _ = it.PeekPtr() {
		it.advance() // INVARIANT: it.cur should never become it.l.tail
	}
	return it
}

// DrainWhile discards next items until it encounters the first item that
// does not match the specified predicate. Modifies the underlying list.
// Returns the iterator itself.
func (it *Iter[T]) DrainWhile(pred func(T) bool) *Iter[T] {
	for v, ok := it.Next(); ok && pred(v); v, ok = it.Next() {
		it.Remove()
	}
	return it
}

// DrainWhilePtr discards next items until it encounters the first item that
// does not match the specified predicate. Modifies the underlying list.
// Returns the iterator itself.
func (it *Iter[T]) DrainWhilePtr(pred func(*T) bool) *Iter[T] {
	for v, ok := it.NextPtr(); ok && pred(v); v, ok = it.NextPtr() {
		it.Remove()
	}
	return it
}

func (it *Iter[T]) advance() {
	it.cur = it.cur.next
}

func (it *Iter[T]) stepBack() {
	it.cur = it.cur.prev
}
