# chainyq

<img src="docs/images/logo.png" alt="Project Logo" width="250"/>

Chainyq provides fast, ergonomic queues, lists, and iterators for Go,
with a focus on high and predictable performance. The API
offers rich functionality inspired by collections in Rust and Java,
while prioritizing static dispatch and avoiding errors, panics, and the
complexity that comes with them. Written in Go 1.25 with zero dependencies.

Currently available:
- `Deque[T any]` - a highly optimized cache-friendly segmented array deque:
  - Per-side capacity tuning
  - Configurable pooling
  - Fast end operations, random access, and 3 types of iterators
  - **Likely the fastest general-purpose deque you'll find**
- `List[T any]` - a simple doubly-linked list
  - Fast insertion and deletion in the middle, for cases when you need to do it frequently
  - 1.5-2x faster than container.List on end operations
  - 3 types of iterators
- `Stack[T any]` - a slice-based stack
  - Fast `Push`, `Pop` and `Peek`
  - Top-down iterator
- `Seq[T any]` - a lazy sequence
  - `Filter`/`Map`/`Reduce`/`ForEach`/`GroupBy`/`ToMap`/etc.
  - Can be created from iterators, slices, and just functions
  - Supports lazy infinite sequences
- `SyncDeque[T any]` - synchronized wrapper around `Deque[T]`

`Seq[T]` is the only type with dynamically dispatched operations,
because a function pointer is needed to enable laziness.

General interfaces are available in the root package (`chainyq`) with implementation-specific interfaces located in their respective packages.

All methods are designed to not tolerate nil as a receiver to avoid redundant
nil checks - just call `New()` instead or more specialized constructors
if you need extra customization. Free functions that accept pointers
tolerate nil.


## Installation
```bash
go get github.com/zelr0x/chainyq
```


## Deque[T any]

`chainyq.deque.Deque` is a double-ended queue and it's fast.

### Benchmarks

The following benchmarks highlight common workloads for Deque[T].

#### TL;DR:
- Fastest overall: chainyq.Deque
- Most memory efficient: chainyq.Deque
- Random access: gz slightly faster
- Bursty growth: chainyq wins clearly


| PushBack                   |  ns/op   | B/op | allocs/op |
|----------------------------|---------:|-----:|----------:|
| chainyq.Deque              |    4.265 |    8 |         0 |
| chainyq.Deque_Ensure       |    3.460 |    0 |         0 |
| edwingeng.Deque            |    5.420 |    8 |         0 |
| gammazero.Deque            |    5.889 |   14 |         0 |
| gammazero.Deque_SetBaseCap |    3.885 |    8 |         0 |
| chainyq.List               |   38.640 |   24 |         1 |
| container.List             |   61.730 |   55 |         1 |

| PushFront                  | ns/op    | B/op | allocs/op |
|----------------------------|---------:|-----:|----------:|
| chainyq.Deque              |    3.782 |    8 |         0 |
| edwingeng.Deque            |    4.273 |    8 |         0 |
| gammazero.Deque            |    4.732 |   17 |         0 |
| chainyq.List               |   33.800 |   24 |         1 |
| container.List             |   60.950 |   55 |         1 |

| Random churn (int)         | ns/op    | B/op | allocs/op |
|----------------------------|---------:|-----:|----------:|
| chainyq.Deque              |    9.799 |    0 |         0 |
| edwingeng.Deque            |   10.350 |    0 |         0 |
| gammazero.Deque            |   10.880 |    0 |         0 |
| chainyq.List               |   18.560 |   11 |         0 |
| container.List             |   26.020 |   27 |         0 |

| Random churn (big struct)  | ns/op    | B/op | allocs/op |
|----------------------------|---------:|-----:|----------:|
| chainyq.Deque              |    14.56 |    0 |         0 |
| edwingeng.Deque            |    15.61 |    0 |         0 |
| gammazero.Deque            |    15.06 |    0 |         0 |

| Random access (get by index) | ns/op    | B/op | allocs/op |
|------------------------------|---------:|-----:|----------:|
| chainyq.Deque                |    9.127 |    0 |         0 |
| edwingeng.Deque              |  2758.00 |    0 |         0 |
| gammazero.Deque              |    8.606 |    0 |         0 |
*Ring buffer should win here, but the gain is in the noise range*

| Bursts of 100k writes/10k reads | ns/op    | B/op | allocs/op |
|---------------------------------|---------:|-----:|----------:|
| chainyq.Deque					  |    17.34 |    0 |         0 |
| edwingeng.Deque				  |    19.90 |    0 |         0 |
| gammazero.Deque				  |    36.05 |    0 |         0 |
*Segmented array should win here*

| Sliding window             | ns/op    | B/op | allocs/op |
|----------------------------|---------:|-----:|----------:|
| chainyq.Deque              |    18.16 |    0 |         0 |
| edwingeng.Deque            |    19.95 |    0 |         0 |
| gammazero.Deque            |    17.07 |    0 |         0 |

*Ring buffer should win here, but the gain is minimal*

Across all benchmarks, `chainyq.Deque` has delivered excellent performance:
- On small input sizes, its performance is virtually indistinguishable
from a ring buffer.
- On larger input sizes, it doesn't fall behind a ring buffer even on workloads
that favor them. Random access and sliding window benchmarks demonstrate
this clearly: a segmented array can't beat a good ring buffer there, yet
`chainyq.Deque` keeps up with the fastest.
- On workloads that involve growing it shows the best performance with both small and
large sized items.

Full results are available in bench.txt, the code for benchmarks is at `/deque/deque_benchmark_test.go`.


```go
import (
	"github.com/zelr0x/chainyq/deque"
)
func main() {
	d := deque.FromSlice([]int{2, 4, 8, 16})
	v, _ := d.PopBack()  // 16, true
	d.PushFront(1)
	if v, ok := d.PopFront(); ok {  // v = 1
		fmt.Printf("popped from front: %v\n", v)
	}
	d.ToSlice() // [2 4 8]
	d.Iter().ForEachPtr(func(x *int) bool {
		*x *= 10
		return true // Can stop early here
	})
	d.ToSlice() // [2 40 80]
}
```

`chainyq.deque` package provides constructors `New`/`NewPooled`/`WithCfg`/`NewValue`/`FromSlice`/`FromSliceCfg` and helpers `SuggestBlockSize`/`Len`.

The API of the deque itself consists of:
- General collection operations: `Len`/`IsEmpty`
- Peeks: `Front`/`FrontPtr`/`Back`/`BackPtr`
- Mutations: `PushFront`/`PushBack`/`PopFront`/`PopBack`
- Random access operations: `Get`/`GetPtr`/`Set`
- Memory control: `EnsureFront`/`EnsureBack`/`ShrinkToFit`/`Clear`/`ClearRelease`
- Iterators: `Iter`/`RevIter`/`BidiIter`
- Misc: `String`/`GoString`/`Equals`
- Slicing: `Slice`/`PtrSlice`
- Collecting: `ToSlice`/`ToPtrSlice`/`ToChan`/`ToPtrChan`

See docs for the full API.

Synchronized version of the Deque is available as `chainyq.deque.syncdeque.SyncDeque`.


## List[T any]

`list.List` is a doubly-linked list that can do all the operations you expect from a linked list and more,
giving you full control without having to manipulate individual nodes and risking making a subtle mistake.
The list is not safe for concurrent use.

```go
import (
	"github.com/zelr0x/chainyq/list"
)

func main() {
	l := list.New[int]()
	l.Add(2).Add(4)
	l.PushBack(8)
	l.PushFront(1)
	l.String() // -> List[1, 2, 4, 8]

	l = list.FromSlice[int]([]int{1, 2, 4, 8, 4, 2, 1})
	eq := func(a, b int) bool { return a == b }
	idx := l.IndexOf(2, eq)    // -> 1
	idx = l.LastIndexOf(2, eq) // -> 5
	idx = l.IndexOf(65535, eq) // -> -1
}
```


## Iterators

The iterators provide statically dispatched traversal and some operations
on the not-yet-traversed items.
They are optimized for every particular data structure - for example,
`Skip` for the `List` will traverse individual nodes one by one,
but for the `Deque` it will just calculate the block and offset and jump directly to it.

The iterators also take into account the semantics of each data structure:
- `Stack` only has a top-to-bottom iterator
- `Deque` and `List` have three types of iterators - `Iter`, `RevIter` and `BidiIter`.

### Iterator API

- Per-item traversal: `HasNext`, `Next`, `NextPtr`, `Peek`, `PeekPtr`
- Skipping: `Skip`, `SkipWhile`, `SkipWhilePtr`
- Position reset: `Reset`
- Actions on remaining items: `ForEach`, `ForEachPtr`
- Collecting remaining items: `ToChan`, `ToPtrChan`, `TakeSlice`, `TakePtrSlice`, `TakeWhile`, `TakeWhilePtr`
- Conversions: `Clone`, `Seq`, `PtrSeq`

`BidiIter` also has:
- Per-item traversal: `HasPrev`, `Prev`, `PrevPtr`, `PeekBack`, `PeekBackPtr`
- Skipping: `SkipBack`
- Position reset: `ResetBack`
- Conversions: `RevSeq`, `RevPtrSeq`

In bidirectional data structures, `Iter` and `RevIter` can be converted to
the opposite direction with `Rev` and to `BidiIter` with `Bidi`.

Methods on `BidiIter` have the same traversal direction as `Iter`, except for special
reverse-direction methods that exist only on `BidiIter`.

`list.List` iterators also have methods for item insertion and removal.

```go
import (
	"github.com/zelr0x/chainyq/list"
)

func main() {
	l = list.FromSlice[int]([]int{1, 2, 4, 8, 4, 2, 1})

	l.Iter().Skip(2).ForEachPtr(func(v *int) bool {
		*v *= 2
		return true
	})
	l.String() // -> List[1 2 8 16 8 4 2]

	it := l.BidiIter()
	v, ok := it.Prev() // -> 0, false
	v, ok = it.Peek() // -> 1, true
	v, ok = it.Next() // -> 1, true
	ok = it.InsertBefore(55) // -> true
	v, ok = it.Prev() // -> 55, true
	v, ok = it.Remove() // -> 55, true
	v, ok = it.Next() // -> 1, true
	v, ok = it.PeekBack() // 0, false

	l.String() // -> List[1 2 8 16 8 4 2]

	v, ok = it.ResetBack().PrevPtr() // -> int* pointing to 2, ok
	*v += 98
	ok = it.InsertAfter(42)
	l.String() // -> List[1 2 8 16 8 4 100 42]

	it.Reset() // -> move to head
		SkipWhile(func(v int) bool {
			return v < 16
		}).  // it.Next() would return 16 here
		SkipBack(1). // now it.Next() would return 8
		DrainWhile(func(v int) bool {
			return v != 4 // skip over [8 16 8]
		})
	slice1 := l.ToSlice() // [1 2 4 100 42]
	slice2 := it.Reset().Skip(1).TakeSlice(3) // [2 4 100]
}
```
See more in the documentation.


## Stack[T any]
`stack.Stack` is a slice-based stack.

```go
import (
	"github.com/zelr0x/chainyq/list"
)

func main() {
	s := stack.New[int]()
	s.Push(1)
	s.Push(2)
	if v, ok := s.PeekPtr(); ok {
		*v = 50
	}
	s.ToSlice() // [50 2 1]
	s.UnwrapCopy() // [1 2 50] (new slice)
	s.UnwrapUnsafe() // [1 2 50] (access underlying slice directly)
}
```


## Seq[T any]
`seq.Seq` is an extremely lightweight lazy, composable sequence abstraction.
It wraps a generator function `func() (T, bool)` and exposes a fluent API
for filtering, transforming, grouping, and collecting values. Unlike eager
collections, Seq only computes elements on demand, making it efficient for
pipelines and transformations over potentially large or infinite sources.

`Seq` is heavily inspired by functional programming but it diverges from
many implementations in that `Seq` does not bring a plethora of special types
that do the heavy lifting - `Seq` is a very thin wrapper around a closure.
That's it. A single function pointer and whatever state the closure captures.
Extremely simple, lightweight, zero dependencies.

```go
import (
	"github.com/zelr0x/chainyq/seq"
)

func main() {
	slice := []string{"apple", "apricot", "banana", "cake"}
	s := seq.FromSlice(slice).Filter(func(s string) bool {
		return s[0] != 'c'
	})
	m := seq.GroupBy(s, func(s string) string {
		return string(s[0])
	})
	m  // {"a": ["apple", "apricot"], "b": "banana"}
}
```

### Seq API

`Seq` can be created from a function with `seq.New`,
from a slice with `seq.FromSlice`, or from a channel with `seq.FromChan`.
A nil-safe zero value for `Seq` can be created with `seq.Empty[T]()`.

All operations on sequences can be divided into two subcategories:
- Intermediate - the ones returning `Seq`
- Terminating - the ones returning any other type.

#### Intermediate operations

`Seq[T]` has a `Next` method just as iterators do.
It also has `Skip` and `SkipWhile`, which return a new `Seq[T]`,
starting from the next element after the skipped ones.
Think of all three as consuming operations: once items are traversed,
they are no longer part of the sequence.

`Seq[T]` also has transforming and filtering operations: `Seq.Filter[T]`,
`Seq.Take[T]`, `seq.Map[T, R]`, and `seq.FlatMap[T, R]`.
The latter two are not methods, but functions in the `seq` package.
That is because Go methods currently can't introduce new generic
parameters - this limits the fluency of the API, but worry not - that
will change when generic methods finally arrive.
For now though, as a rule of thumb - operations that leave the sequence's
element type unchanged are defined as methods on `seq.Seq` and
operations that transform the element type are functions in the `seq` package.

#### Terminating operations

You can call an arbitrary function on all items in the sequence using
`ForEach`, `ForEachUntil`, and `ForEachIndexed` - with each having
finer-grained control than the previous one. These operations are
used to cause side-effects, they don't produce new sequences and don't
collect items, unless you explicitly do so inside a closure you pass
as an argument.

You can collect items using `Seq.ToSlice[T]`, `Seq.ToChan[T]`,
`seq.ToMap[T, K, V]`, `seq.ToMapMerge[T, K, V]` or some of their versions
that allow specifying capacity hints.

You can also group items into categories using `seq.GroupBy[T, K]`.

See docs for the full list of available operations.

### More complex example

The next example is intentionally overengineered to demonstrate how `Seq`
integrates seamlessly with other Chainyq collections, enabling complex
pipelines. While this particular case is simple, the same pattern becomes
much more useful in advanced tasks.
```go
import (
	"github.com/zelr0x/chainyq/seq"
	"github.com/zelr0x/chainyq/deque"
)

func main() {
	slice := []string{"apple", "apricot", "banana", "cake"}
	s := seq.FromSlice(slice)
	d := deque.FromSlice(seq.Map(s, func(x string) {
		return len(x)
	}).ToSlice())
	d.PopBack()
	sliceByEvenOdd := evenOdd(d.Iter())
}

func evenOdd[T ~int](n chainyq.Nexter[T]) map[bool][]T {
	s := seq.New(n.Next)
	return seq.GroupBy(s, func(x T) bool {
		return x%2 == 0
	})
}
```
