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
package seq

// Seq is a lazy iterator over a potentially infinite sequence of values.
// It is a value type, there's no reason to pass it by reference.
// Seq is not concurrency-safe.
type Seq[T any] struct {
	next func() (T, bool)
}

// New creates a new Seq using the specified generator function.
// The generator should return the next element and a boolean indicating
// whether more elements are available.
//
// Example: finite seq from slice
//
//   xs := []int{1, 2, 3}
//   i := 0
//   s := seq.New(func() (int, bool) {
//   	if i >= len(xs) {
//   		return 0, false
//   	}
//   	v := xs[i]
//   	i++
//   	return v, true
//   })
//   s.ToSlice()  // []int{1, 2, 3}
//
// Example: infinite seq from a counter variable
//
//   i := 0
//   s := seq.New(func() (int, bool) {
//   	i++
//   	return i, true
//   })
//   s.Take(5).ToSlice()  // []int{1, 2, 3, 4, 5}
//
// Example: infinite seq from RNG
//
//   s := seq.New(func() (int, bool) {
//   	return rand.Intn(100), true
//   })
//   s.Take(5).ToSlice()  // []int{42, 7, 88, 13, 56} (random values)
func New[T any](next func() (T, bool)) Seq[T] {
	return Seq[T]{next: next}
}

// Creates a new Seq from the specified slice.
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3, 4})
func FromSlice[T any](slice []T) Seq[T] {
	if len(slice) == 0 {
		return newTerminated[T]()
	}
	i := 0
	return New(func() (T, bool) {
		if i >= len(slice) {
			var zero T
			return zero, false
		}
		v := slice[i]
		i++
		return v, true
	})
}

// FromChan creates a new Seq that pulls elements from the given channel.
// The sequence ends once the channel is closed.
//
// Example: wrap an existing channel
//
//   ch := make(chan int, 3)
//   ch <- 1; ch <- 2; ch <- 3
//   close(ch)
//   s := seq.FromChan(ch)
//   s.ToSlice()  // []int{1, 2, 3}
//
// Example: use with a producer goroutine
//
//   ch := make(chan string)
//   go func() {
//   	defer close(ch)
//   	for _, s := range []string{"a", "b", "c"} {
//   		ch <- s
//   	}
//   }()
//   s := seq.FromChan(ch)
//   s.ToSlice()  // []string{"a", "b", "c"}
func FromChan[T any](ch <-chan T) Seq[T] {
    return New(func() (T, bool) {
        v, ok := <-ch
        return v, ok
    })
}

func newTerminated[T any]() Seq[T] {
	return New(func() (T, bool) {
		var zero T
		return zero, false
	})
}

// Next yields the next element of a sequence.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2})
//   s.Next()  // 1, true
//   s.Next()  // 2, true
//   s.Next()  // 0, false
func (seq Seq[T]) Next() (T, bool) {
	return seq.next()
}

// Filter yields only those elements of the sequence
// for which pred returns true.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3, 4})
//   evens := s.Filter(func(x int) bool { return x % 2 == 0 })
//   evens.ToSlice()  // []int{2, 4}
func (seq Seq[T]) Filter(pred func(T) bool) Seq[T] {
	return New(func() (T, bool) {
		for v, ok := seq.next(); ok; v, ok = seq.next() {
			if pred(v) {
				return v, true
			}
		}
		var zero T
		return zero, false
	})
}

// Take yields at most n elements from the sequence,
// stopping once n elements have been produced or the sequence ends.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3, 4})
//   firstTwo := s.Take(2).ToSlice()  // []int{1, 2}
func (seq Seq[T]) Take(n int) Seq[T] {
	if n < 1 {
		return newTerminated[T]()
	}
	count := 0
	return New(func() (T, bool) {
		if count >= n {
			var zero T
			return zero, false
		}
		v, ok := seq.next()
		if !ok {
			var zero T
			return zero, false
		}
		count++
		return v, true
	})
}

// Skip discards the first n elements of the sequence,
// yielding the remainder.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3, 4})
//   rest := s.Skip(2).ToSlice()  // []int{3, 4}
func (seq Seq[T]) Skip(n int) Seq[T] {
	count := 0
	return New(func() (T, bool) {
		for v, ok := seq.next(); ok; v, ok = seq.next() {
			if count < n {
				count++
				continue
			}
			return v, true
		}
		var zero T
		return zero, false
	})
}

// SkipWhile skips elements from the sequence as long as pred returns true.
// Once pred returns false, the remaining elements are yielded.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3, 4})
//   rest := s.SkipWhile(func(x int) bool { return x < 3 })
//   rest.ToSlice()  // []int{3, 4}
func (seq Seq[T]) SkipWhile(pred func(T) bool) Seq[T] {
	return New(func() (T, bool) {
		for v, ok := seq.next(); ok; v, ok = seq.next() {
			if pred(v) {
				continue
			}
			return v, true
		}
		var zero T
		return zero, false
	})
}

// Map transforms each element of the sequence using f,
// producing a new sequence of type R.
//
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3})
//   doubled := Map(s, func(x int) int { return x * 2 })
//   doubled.ToSlice()  // []int{2, 4, 6}
func Map[T any, R any](seq Seq[T], f func(T) R) Seq[R] {
	return New(func() (R, bool) {
		v, ok := seq.next()
		if !ok {
			var zero R
			return zero, false
		}
		return f(v), true
	})
}

// Count iterates over all the items in the sequence and counts them.
// Returns the count on finite sequences, hangs on infinite sequences.
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3}).Count()  // 3
func (seq Seq[T]) Count() int {
	var count int
	seq.ForEach(func(x T) {
		count++
	})
	return count
}

// Count iterates over all the items in the sequence and counts them.
// Returns the count on finite sequences, hangs on infinite sequences.
// Example:
//
//   s := seq.FromSlice([]int{1, 2, 3}).CountU64()  // 3 (uint64)
func (seq Seq[T]) CountU64() uint64 {
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
//   seq.FromSlice([]int{1, 2, 3}).
// 		ForEach(func(x int) {
//      	fmt.Println(x)
//   	})
func (seq Seq[T]) ForEach(f func(T)) {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		f(v)
	}
}

// ForEachUntil applies f to each element until f returns false,
// at which point iteration stops.
//
// Example:
//
//   seq.FromSlice([]int{1, 2, 3, 4}).
//   	ForEachUntil(func(x int) bool {
//      	fmt.Println(x)
//      	return x < 3 // stop once x >= 3, 3 will be printed
//   	})
func (seq Seq[T]) ForEachUntil(f func(T) bool) {
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if !f(v) {
			return
		}
	}
}

// ForEachIndexed applies f to each element along with its index.
// Iteration stops early if f returns false.
//
// Example:
//
//   seq.FromSlice([]string{"a", "b", "c"}).
//   	ForEachIndexed(func(i int, s string) bool {
//   		fmt.Printf("%d: %s\n", i, s)
//   		return true
//   	})
func (seq Seq[T]) ForEachIndexed(f func(int, T) bool) {
	i := 0;
	for v, ok := seq.next(); ok; v, ok = seq.next() {
		if !f(i, v) {
			return
		}
		i++
	}
}

// ToSlice collects all elements of the sequence into a slice.
//
// Example:
//
//   seq.FromSlice([]int{1, 2, 3}).ToSlice()  // []int{1, 2, 3}
func (seq Seq[T]) ToSlice() []T {
	res := make([]T, 0, 16)
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
//   ch := seq.FromSlice([]int{1, 2, 3}).ToChan(0)
//   for x := range ch {
//   	fmt.Println(x)
//   }
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
//   s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//   m := seq.ToMap(s,
//   	func(s string) string { return string(s[0]) },
//   	func(s string) string { return s },
//   )
//   m["a"]  // "apricot", since it is after "apple" in the sequence source
//   m["b"]  // "banana"
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
//   s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//   m := seq.ToMapHint(s,
//   	func(s string) string { return string(s[0]) },
//   	func(s string) int { return len(s) },
//   	2,
//   )
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
//   s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//   m := seq.ToMapMerge(s,
//   	func(s string) string { return string(s[0]) }, // key: first letter
//   	func(s string) int { return len(s) },          // value: length
//   	func(a, b int) int { return a },               // mergeFn: prefer first
//   	2,
//   )
//
//   // m["a"] = 10 (5+5), m["b"] = 6
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
//   s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//   groups := seq.GroupBy(s, func(s string) string { return string(s[0]) })
//
//   // groups["a"] = []string{"apple", "apricot"}
//   // groups["b"] = []string{"banana"}
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
//   s := seq.FromSlice([]string{"apple", "apricot", "banana"})
//   groups := seq.GroupByHint(s,
//   	func(s string) string { return string(s[0]) },
//   	2, // expect ~2 distinct keys
//   	2, // expect ~2 values per key
//   )
//
//   a := groups["a"]  // []string{"apple", "apricot"}
//   b := groups["b"]  // []string{"banana"}
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
//   s := seq.FromSlice([]int{1, 2, 3})
//   product := seq.Reduce(s, 1, func(acc, x int) int {
//   	return acc * x
//   })
//   // product = 6
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
//   s := seq.FromSlice([]int{1, 2, 3})
//   total := 0
//   ReducePtr(s, &total, func(acc *int, x int) {
//   	*acc += x
//   })
//   // total = 6
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
//   s := seq.FromSlice([]int{1, 2, 3})
//   sum := seq.Fold(s, 0, func(acc, x int) int {
//   	return acc + x
//   })
//   // sum = 6
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
//   s := seq.FromSlice([]int{1, 2, 3})
//   sum := seq.Foldr(s, 0, func(x int, rest func() int) int {
//   	return x + rest()
//   })
//   sum  // 6
//
// Example: subtraction is not associative, so results will be different from Fold
//   s = seq.FromSlice([]int{1, 2, 3})
//   res = seq.Foldr(s, 0, func(x int, rest func() int) int {
//   	return x - rest()
//   })
//   res  // 2, because it is evaluated as 1 - (2 - (3 - 0)) = 2
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
