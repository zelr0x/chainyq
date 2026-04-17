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

// Package seq defines a lazy sequence that is also often called a stream
// or an iterator. Most operations are defined as methods, but some have
// to be defined as free functions, primarily due to go's generic limitations.
package seq

import (
	"iter"
	"math"

	"github.com/zelr0x/chainyq/internal/iterutil"
	"github.com/zelr0x/chainyq/internal/numutil"
)

// Seq is a lazy iterator over a potentially infinite sequence of values.
// No elements are consumed until the downstream is iterated.
//
// Transformations on a sequence (upstream) produce a new sequence
// (downstream), which shares the underlying source with the upstream.
// The items consumed from downstream sequences are consumed from upstream
// all sequences unless different behavior is explicitly documented.
//
// As a rule of thumb: after applying a transformation, do not continue using
// the original sequence unless it is explicitly specified to be safe.
// For example, it is safe to use the upstream after [Take] or [Skip] - it will
// contain the remaining items. As another example, [Count] will not consume
// any items at all if exact size of the sequence is known.
// [IsExactSized] can be used to check if exact size of the sequence is known.
//
// Seq is a value type, there's no reason to pass it by reference.
//
// Seq is not concurrency-safe.
type Seq[T any] struct {
	next      func() (T, bool)
	skip      func(int)
	remaining *uint64
	maxLen    uint64
}

// New creates a new Seq using the specified generator function.
// The generator should return the next element and a boolean indicating
// whether more elements are available.
//
// Example: finite seq from slice
//
//	xs := []int{1, 2, 3}
//	i := 0
//	s := seq.New(func() (int, bool) {
//		if i >= len(xs) {
//			return 0, false
//		}
//		v := xs[i]
//		i++
//		return v, true
//	})
//	s.ToSlice()  // []int{1, 2, 3}
//
// Example: infinite seq from a counter variable
//
//	i := 0
//	s := seq.New(func() (int, bool) {
//		i++
//		return i, true
//	})
//	s.Take(5).ToSlice()  // []int{1, 2, 3, 4, 5}
//
// Example: infinite seq from RNG
//
//	s := seq.New(func() (int, bool) {
//		return rand.Intn(100), true
//	})
//	s.Take(5).ToSlice()  // []int{42, 7, 88, 13, 56} (random values)
func New[T any](next func() (T, bool)) Seq[T] {
	return newSeq(next, nil, nil, 0)
}

// Sized creates a Seq with a known upper-bound on length.
func Sized[T any](next func() (T, bool), maxLen int) Seq[T] {
	if maxLen <= 0 {
		return newTerminated[T]()
	}
	return SizedU64(next, uint64(maxLen))
}

// SizedU64 creates a Seq with a known upper-bound on length.
func SizedU64[T any](next func() (T, bool), maxLen uint64) Seq[T] {
	if maxLen <= 0 {
		return newTerminated[T]()
	}
	return newSeq(next, nil, nil, maxLen)
}

// ExactSized creates a Seq with a known exact length.
func ExactSized[T any](next func() (T, bool), exactLen int) Seq[T] {
	if exactLen <= 0 {
		return newTerminated[T]()
	}
	return ExactSizedU64(next, uint64(exactLen))
}

// ExactSizedU64 creates a Seq with a known exact length.
func ExactSizedU64[T any](next func() (T, bool), exactLen uint64) Seq[T] {
	if exactLen <= 0 {
		return newTerminated[T]()
	}
	return newSeq(next, nil, &exactLen, exactLen)
}

func (seq Seq[T]) WithSkip(skip func(int)) Seq[T] {
	seq.skip = skip
	return seq
}

// Creates a new Seq from the specified slice.
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3, 4})
func FromSlice[T any](slice []T) Seq[T] {
	n := len(slice)
	if n == 0 {
		return newTerminated[T]()
	}
	i := 0
	return ExactSized(func() (T, bool) {
		if i >= n {
			var zero T
			return zero, false
		}
		v := slice[i]
		i++
		return v, true
	}, n).WithSkip(func(k int) {
		i += k
	})
}

// FromChan creates a new Seq that pulls elements from the given channel.
// The sequence ends once the channel is closed.
//
// Example: wrap an existing channel
//
//	ch := make(chan int, 3)
//	ch <- 1; ch <- 2; ch <- 3
//	close(ch)
//	s := seq.FromChan(ch)
//	s.ToSlice()  // []int{1, 2, 3}
//
// Example: use with a producer goroutine
//
//	ch := make(chan string)
//	go func() {
//		defer close(ch)
//		for _, s := range []string{"a", "b", "c"} {
//			ch <- s
//		}
//	}()
//	s := seq.FromChan(ch)
//	s.ToSlice()  // []string{"a", "b", "c"}
func FromChan[T any](ch <-chan T) Seq[T] {
	return New(func() (T, bool) {
		v, ok := <-ch
		return v, ok
	})
}

// Empty creates an empty Seq. This is the same as var s Seq[T]
// but nil-safe and clearly documents the intent.
func Empty[T any]() Seq[T] {
	return newTerminated[T]()
}

func newTerminated[T any]() Seq[T] {
	// It is possible to use ExactSized here, but it will require allocating
	// a pointer to 0. Shared ptr is potentially brittle, so it is unsized for now.
	return New(func() (T, bool) {
		var zero T
		return zero, false
	})
}

func newSeq[T any](
	next func() (T, bool),
	skip func(int),
	remaining *uint64,
	maxLen uint64,
) Seq[T] {
	return Seq[T]{next: next, skip: skip, remaining: remaining, maxLen: maxLen}
}

func (seq Seq[T]) lenOr(fallback int) int {
	if seq.IsExactSized() {
		return numutil.U64ToInt(*seq.remaining)
	}
	if seq.maxLen > 0 {
		return numutil.U64ToInt(seq.maxLen)
	}
	return fallback
}

// IsExactSized returns true if the sequence is ExactSized
// i.e. it knows the exact size of the underlying source.
func (seq Seq[T]) IsExactSized() bool {
	// This implementation itself is currently a contract relied upon.
	return seq.remaining != nil
}

func (seq Seq[T]) decRemaining(n uint64) {
	if n <= 0 {
		return
	}
	if seq.remaining != nil {
		rem := *seq.remaining
		if rem <= n {
			rem = 0
		} else {
			rem -= n
		}
		*seq.remaining = rem
	}
}

// Next yields the next element of a sequence.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2})
//	s.Next()  // 1, true
//	s.Next()  // 2, true
//	s.Next()  // 0, false
func (seq Seq[T]) Next() (T, bool) {
	v, ok := seq.next()
	if ok {
		seq.decRemaining(1)
	}
	return v, ok
}

// Filter creates a new sequence that yields only those upstream items
// that satisfy the given predicate, fully consuming the upstream
// in the process.
//
// Resulting sequence loses ExactSized capability if upstream had it.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3, 4})
//	evens := s.Filter(func(x int) bool { return x % 2 == 0 })
//	evens.ToSlice()  // []int{2, 4}
func (seq Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return newSeq(func() (T, bool) {
		for v, ok := seq.next(); ok; v, ok = seq.next() {
			if pred(v) {
				return v, true
			}
		}
		if seq.IsExactSized() {
			*seq.remaining = 0
		}
		var zero T
		return zero, false
	}, nil, nil, seq.maxLen)
}

// Take yields at most n elements from the sequence,
// stopping once n elements have been produced or the sequence ends.
//
// Does not invalidate upstream, only consumes the specified number
// of items from it.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3, 4})
//	firstTwo := s.Take(2).ToSlice()  // []int{1, 2}
func (seq Seq[T]) Take(n int) Seq[T] {
	if n <= 0 {
		return newTerminated[T]()
	}
	nu64 := uint64(n)
	var rem *uint64 = nil
	if seq.IsExactSized() {
		r := *seq.remaining
		if r < nu64 {
			nu64 = r
		}
		rem = &nu64
	}
	var i uint64 = 0
	return newSeq(func() (T, bool) {
		if nu64 <= 0 {
			seq.decRemaining(i)
			var zero T
			return zero, false
		}
		v, ok := seq.next()
		i++
		if !ok {
			seq.decRemaining(i)
			var zero T
			return zero, false
		}
		nu64--
		return v, true
	}, seq.skip, rem, nu64)
}

// TakeWhile returns a new Seq that yields elements from upstream
// as long as the given predicate returns true. Once the predicate
// fails, iteration stops permanently.
//
// Does not invalidate upstream, only consumes items from it
// while predicate is satisfied, plus the first item that
// fails to satisfy the predicate.
func (seq Seq[T]) TakeWhile(pred func(T) bool) Seq[T] {
	done := false
	var i uint64 = 0
	return newSeq(func() (T, bool) {
		if done {
			var zero T
			return zero, false
		}
		v, ok := seq.next()
		if !ok {
			done = true
			seq.decRemaining(i)
			var zero T
			return zero, false
		}
		i++
		if pred(v) {
			return v, true
		}
		done = true
		seq.decRemaining(i)
		var zero T
		return zero, false
	}, seq.skip, seq.remaining, seq.maxLen)
}

// Skip discards the first n elements of the sequence,
// yielding the remainder.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3, 4})
//	rest := s.Skip(2).ToSlice()  // []int{3, 4}
func (seq Seq[T]) Skip(n int) Seq[T] {
	if n <= 0 {
		return seq
	}
	nu64 := uint64(n)
	if seq.maxLen > 0 && nu64 > seq.maxLen {
		return newTerminated[T]()
	}
	var rem *uint64 = nil
	maxLen := seq.maxLen
	if seq.IsExactSized() {
		r := *seq.remaining
		if r <= nu64 {
			return newTerminated[T]()
		}
		r -= nu64
		rem = &r
		maxLen = r
	} else if maxLen > nu64 {
		maxLen -= nu64
	}
	return newSeq(func() (T, bool) {
		if n > 0 {
			if seq.skip != nil {
				seq.skip(n)
			} else {
				for range n {
					_, _ = seq.next()
				}
			}
			n = 0
		}
		return seq.next()
	}, seq.skip, rem, maxLen)
}

// SkipWhile skips elements from the sequence as long as pred returns true.
// Once pred returns false, the remaining elements are yielded.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3, 4})
//	rest := s.SkipWhile(func(x int) bool { return x < 3 })
//	rest.ToSlice()  // []int{3, 4}
func (seq Seq[T]) SkipWhile(pred func(T) bool) Seq[T] {
	skipped := false
	return newSeq(func() (T, bool) {
		if !skipped {
			var i uint64 = 0
			for v, ok := seq.next(); ok; v, ok = seq.next() {
				if pred(v) {
					i++
					continue
				}
				skipped = true
				seq.decRemaining(i)
				return v, ok
			}
		}
		return seq.next()
	}, seq.skip, seq.remaining, seq.maxLen)
}

// Map transforms each element of the sequence using f,
// producing a new sequence of type R.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	doubled := Map(s, func(x int) int { return x * 2 })
//	doubled.ToSlice()  // []int{2, 4, 6}
func Map[T any, R any](seq Seq[T], f func(T) R) Seq[R] {
	return newSeq(func() (R, bool) {
		v, ok := seq.next()
		if !ok {
			var zero R
			return zero, false
		}
		return f(v), true
	}, seq.skip, seq.remaining, seq.maxLen)
}

// FlatMap transforms each element of the sequence using the given function f,
// which itself produces a subsequence, including each element of each such
// subsequence into the resulting sequence.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	flat := seq.FlatMap(s, func(x int) seq.Seq[int] {
//		return seq.FromSlice([]int{x, x * 10})
//	}).ToSlice()  // [1 10 2 20 3 30]
func FlatMap[T any, R any](seq Seq[T], f func(T) Seq[R]) Seq[R] {
	inner := Empty[R]()
	return newSeq(func() (R, bool) {
		for {
			v, ok := inner.next()
			if ok {
				return v, ok
			}
			outer, ok := seq.next()
			if !ok {
				var zero R
				return zero, false
			}
			inner = f(outer)
		}
	}, nil, nil, 0) // nil and 0 because FlatMap can grow past maxLen.
}

// Find returns the first element in the sequence that satisfies
// the given predicate. It short-circuits as soon as a match is found.
// Returns zero and false if no element satisfies the predicate.
func (seq Seq[T]) Find(pred func(T) bool) (T, bool) {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// All tests if all items in the sequence satisfy the given predicate.
// Returns true if all elements satisfy the predicate.
// All is short-circuiting; it will stop processing as soon as it finds
// an item that does not satisfy the given predicate.
// Returns true on empty sequences because in this case there is no item
// that does not satisfy the predicate.
func (seq Seq[T]) All(pred func(T) bool) bool {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if !pred(v) {
			return false
		}
	}
	return true
}

// Any tests if any item in the sequence satisfies the given predicate.
// Returns true if any item satisfies the predicate.
// Any is short-circuiting; it will stop processing as soon as it finds
// an item that satisfies the given predicate.
// Returns false on empty sequences, because in this case there is no item
// that satisfies the predicate.
func (seq Seq[T]) Any(pred func(T) bool) bool {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if pred(v) {
			return true
		}
	}
	return false
}

// None tests if no item in the sequence satisfies the given predicate.
// Returns true if no element satisfies the predicate.
// None is short-circuiting; it will stop processing as soon as it finds
// an item that satisfies the given predicate.
// Returns true on empty sequences, because in this case there is no item
// that satisfies the predicate.
func (seq Seq[T]) None(pred func(T) bool) bool {
	return !seq.Any(pred)
}

// Count consumes all items in the sequence and returns their count.
// Hangs forever on infinite sequences.
//
// Does not consume ExactSized sequences.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3}).Count()  // 3
func (seq Seq[T]) Count() int {
	if seq.IsExactSized() {
		return numutil.U64ToInt(*seq.remaining)
	}
	var count int
	seq.ForEach(func(x T) {
		count++
	})
	if count < 0 {
		return math.MaxInt
	}
	return count
}

// Count consumes all items in the sequence and returns their count.
// Hangs forever on infinite sequences.
//
// Does not consume ExactSized sequences.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3}).CountU64()  // 3 (uint64)
func (seq Seq[T]) CountU64() uint64 {
	if seq.IsExactSized() {
		return *seq.remaining
	}
	var count uint64
	seq.ForEach(func(x T) {
		count++
	})
	return count
}

// ForEach applies f to each element of the sequence.
// Iteration always runs through all elements.
//
// Example:
//
//	  seq.FromSlice([]int{1, 2, 3}).
//			ForEach(func(x int) {
//	     	fmt.Println(x)
//	  	})
func (seq Seq[T]) ForEach(f func(T)) {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		f(v)
	}
	if seq.IsExactSized() {
		*seq.remaining = 0
	}
}

// ForEachUntil applies f to each element until f returns false,
// at which point iteration stops.
//
// Example:
//
//	seq.FromSlice([]int{1, 2, 3, 4}).
//		ForEachUntil(func(x int) bool {
//	   	fmt.Println(x)
//	   	return x < 3 // stop once x >= 3, 3 will be printed
//		})
func (seq Seq[T]) ForEachUntil(f func(T) bool) {
	var i uint64 = 0
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if !f(v) {
			seq.decRemaining(i)
			return
		}
		i++
	}
	if seq.IsExactSized() {
		*seq.remaining = 0
	}
}

// ForEachIndexed applies f to each element along with its index.
// Iteration stops early if f returns false.
//
// Example:
//
//	seq.FromSlice([]string{"a", "b", "c"}).
//		ForEachIndexed(func(i int, s string) bool {
//			fmt.Printf("%d: %s\n", i, s)
//			return true
//		})
func (seq Seq[T]) ForEachIndexed(f func(int, T) bool) {
	i := 0
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if !f(i, v) {
			if seq.IsExactSized() {
				seq.decRemaining(uint64(i))
			}
			return
		}
		i++
	}
	if seq.IsExactSized() {
		*seq.remaining = 0
	}
}

// Slice takes a subsequence of size end-start items after skipping start items.
// If start or end are negative or if end is greater than start, empty sequence
// is returned.
func (s Seq[T]) Slice(start, end int) Seq[T] {
	if start < 0 || end <= start {
		return newTerminated[T]()
	}
	return s.Skip(start).Take(end - start)
}

// IterSlice takes a slice of size end-start items after skipping start items
// and returns it as [iter.Seq], so Seq can be iterated with range.
// If start or end are negative or if end is greater than start, empty sequence
// is returned.
func (s Seq[T]) IterSlice(start, end int) iter.Seq[T] {
	return s.Slice(start, end).IterAll()
}

// IterAll returns [iter.Seq] so Seq can be iterated with range.
func (s Seq[T]) IterAll() iter.Seq[T] {
	return iterutil.IterSeq(s.next)
}

// ToSlice collects all elements of the sequence into a slice.
//
// Example:
//
//	seq.FromSlice([]int{1, 2, 3}).ToSlice()  // []int{1, 2, 3}
func (seq Seq[T]) ToSlice() []T {
	res := make([]T, 0, seq.lenOr(16))
	seq.ForEach(func(x T) {
		res = append(res, x)
	})
	return res
}

// ToChan drains the sequence into a channel of size `size`.
// The channel is closed once all elements are sent.
//
// Example:
//
//	ch := seq.FromSlice([]int{1, 2, 3}).ToChan(0)
//	for x := range ch {
//		fmt.Println(x)
//	}
func (seq Seq[T]) ToChan(size int) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		for v, ok := seq.next(); ok; v, ok = seq.next() {
			ch <- v
		}
	}()
	return ch
}

// ToMap collects elements of a sequence into a map[K]V.
// keyFn extracts the key from each element.
// valFn maps each element to a value.
// If multiple elements produce the same key, the last one wins.
//
// Example:
//
//	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//	m := seq.ToMap(s,
//		func(s string) string { return string(s[0]) },
//		func(s string) string { return s },
//	)
//	m["a"]  // "apricot", since it is after "apple" in the sequence source
//	m["b"]  // "banana"
func ToMap[T any, K comparable, V any](
	seq Seq[T],
	keyFn func(T) K,
	valFn func(T) V,
) map[K]V {
	return ToMapHint(seq, keyFn, valFn, 16)
}

// ToMapHint is like ToMap but allows specifying an expected number of
// distinct keys to preallocate the map. This reduces rehashing overhead
// when the number of keys is known or can be estimated.
//
// nKeysHint is a capacity hint for the map.
//
// Example:
//
//	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//	m := seq.ToMapHint(s,
//		func(s string) string { return string(s[0]) },
//		func(s string) int { return len(s) },
//		2,
//	)
func ToMapHint[T any, K comparable, V any](
	seq Seq[T],
	keyFn func(T) K,
	valFn func(T) V,
	nKeysHint int,
) map[K]V {
	res := make(map[K]V, nKeysHint)
	seq.ForEach(func(x T) {
		k := keyFn(x)
		v := valFn(x)
		res[k] = v
	})
	return res
}

// ToMapMerge collects elements of a sequence into a map[K]V with conflict
// resolution. keyFn extracts the key, valFn maps each element to a value,
// and mergeFn combines two values when the same key appears multiple times.
//
// nKeysHint is a capacity hint for the map.
//
// Example: prefer the first value with the same key on conflict
//
//	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//	m := seq.ToMapMerge(s,
//		func(s string) string { return string(s[0]) }, // key: first letter
//		func(s string) int { return len(s) },          // value: length
//		func(a, b int) int { return a },               // mergeFn: prefer first
//		2,
//	)
//
//	// m["a"] = 10 (5+5), m["b"] = 6
func ToMapMerge[T any, K comparable, V any](
	seq Seq[T],
	keyFn func(T) K,
	valFn func(T) V,
	mergeFn func(V, V) V,
	nKeysHint int,
) map[K]V {
	res := make(map[K]V, nKeysHint)
	seq.ForEach(func(x T) {
		k := keyFn(x)
		v := valFn(x)
		if existing, ok := res[k]; ok {
			res[k] = mergeFn(existing, v)
		} else {
			res[k] = v
		}
	})
	return res
}

// GroupBy collects elements of a sequence into slices keyed by the result
// of keyFn. It returns a map[K][]T where each key corresponds to a bucket
// of values. This is a convenience wrapper around [GroupByHint] with default
// allocation hints.
//
// Example:
//
//	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//	groups := seq.GroupBy(s, func(s string) string { return string(s[0]) })
//
//	// groups["a"] = []string{"apple", "apricot"}
//	// groups["b"] = []string{"banana"}
func GroupBy[T any, K comparable](
	seq Seq[T],
	keyFn func(T) K,
) map[K][]T {
	return GroupByHint(seq, keyFn, 8, 8)
}

// GroupByHint collects elements of a sequence into slices keyed by the
// result of keyFn. It returns a map[K][]T where each key corresponds to
// a bucket of values.
//
// The nKeysHint parameter is used to preallocate the map with an expected
// number of distinct keys, reducing rehashing overhead. The nValsHint
// parameter is used to preallocate each slice with an expected number of
// values per key, reducing repeated allocations as values are appended.
//
// Example:
//
//	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//	groups := seq.GroupByHint(s,
//		func(s string) string { return string(s[0]) },
//		2, // expect ~2 distinct keys
//		2, // expect ~2 values per key
//	)
//
//	a := groups["a"]  // []string{"apple", "apricot"}
//	b := groups["b"]  // []string{"banana"}
//
// If nKeysHint or nValsHint are set too low, the function still works
// correctly; they only affect allocation efficiency.
func GroupByHint[T any, K comparable](
	seq Seq[T],
	keyFn func(T) K,
	nKeysHint int,
	nValsHint int,
) map[K][]T {
	res := make(map[K][]T, nKeysHint)
	seq.ForEach(func(x T) {
		k := keyFn(x)
		if _, ok := res[k]; !ok {
			res[k] = make([]T, 0, nValsHint)
		}
		res[k] = append(res[k], x)
	})
	return res
}

// Reduce performs a left reduction over the sequence.
// It is equivalent to Fold, taking an initial accumulator and a
// function that combines it with each element.
//
// acc is the initial accumulator value.
// f takes the current accumulator and the next element.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	product := seq.Reduce(s, 1, func(acc, x int) int {
//		return acc * x
//	})
//	// product = 6
func Reduce[T any, R any](
	seq Seq[T],
	acc R,
	f func(R, T) R,
) R {
	return Fold(seq, acc, f)
}

// ReducePtr performs a left reduction over the sequence, mutating
// the accumulator pointer in place. This is useful when the accumulator
// is a complex type or when in‑place updates are desired.
//
// acc is a pointer to the accumulator value.
// f takes the accumulator pointer and the next element, and mutates acc.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	total := 0
//	ReducePtr(s, &total, func(acc *int, x int) {
//		*acc += x
//	})
//	// total = 6
func ReducePtr[T any, R any](
	seq Seq[T],
	acc *R,
	f func(*R, T),
) {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		f(acc, v)
	}
}

// Fold performs a left fold over the sequence.
// It applies function f to the accumulator and each element in order,
// returning the final accumulated result.
//
// acc is the initial accumulator value.
// f takes the current accumulator and the next element.
//
// Example:
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	sum := seq.Fold(s, 0, func(acc, x int) int {
//		return acc + x
//	})
//	// sum = 6
func Fold[T any, R any](
	seq Seq[T],
	acc R,
	f func(R, T) R,
) R {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		acc = f(acc, v)
	}
	return acc
}

// Foldr performs a right fold over the sequence.
// It applies function f to each element and a thunk (func() R) that
// represents folding the rest of the sequence. This allows lazy or
// short‑circuiting folds.
//
// acc is the initial accumulator value.
// f takes the current element and a function that computes the rest.
//
// Example: addition is associative, so results will be the same as Fold
//
//	s := seq.FromSlice([]int{1, 2, 3})
//	sum := seq.Foldr(s, 0, func(x int, rest func() int) int {
//		return x + rest()
//	})
//	sum  // 6
//
// Example: subtraction is not associative, so results will be different from Fold
//
//	s = seq.FromSlice([]int{1, 2, 3})
//	res = seq.Foldr(s, 0, func(x int, rest func() int) int {
//		return x - rest()
//	})
//	res  // 2, because it is evaluated as 1 - (2 - (3 - 0)) = 2
func Foldr[T any, R any](
	seq Seq[T],
	acc R,
	f func(T, func() R) R,
) R {
	v, ok := seq.next()
	if !ok {
		return acc
	}
	return f(v, func() R {
		return Foldr(seq, acc, f)
	})
}
