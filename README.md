# chainyq

<img src="docs/images/logo.png" alt="Project Logo" width="250"/>

Chainyq provides high-performance collections for go, with a rich API
and a strict design: no dynamic dispatch, panics, or errors.
Written in Go 1.25 with zero-dependencies.

Currently available:
- `Deque[T any]` - a cache-friendly segmented array deque (double-ended queue)
with per-side capacity tuning and configurable pooling.
Supports fast end operations, random access, and iterators.
- `List[T any]` - a simple doubly-linked list that is a deque for cases
when you need frequent insertion and/or deletion in the middle.
- `Stack[T any]` - a slice-based stack with normal `Push`/`Pop`
and top-to-bottom iterator.
- `Seq[T any]` - a lazy sequence with 
`Filter`/`Map`/`Reduce`/`ForEach`/`GroupBy`/`ToMap` kind of operations.
Can be created from iterators, slices, and just functions, which is useful
for creating infinite sequences.
- `SyncDeque[T any]` - wrapper around `Deque[T]` with `sync.RWMutex`.

`Seq[T]` is the only type with dynamically dispatched operations,
because a function pointer is needed to enable laziness.

General interfaces are available in the root package (`chainyq`) with implementation-specific interfaces located in their respective packages.

All methods are designed to not tolerate nil as a receiver to avoid redundant
nil checks - just call `New()` instead or more specialized constructors
if you need extra customization. Free functions that accept pointers
tolerate nil.


## Deque[T any]

`chainyq.deque.Deque` is probably the fastest general-purpose deque you can find:

| PushBack                   |  ns/op   | B/op | allocs/op |
|----------------------------|----------|------|-----------|
| chainyq.Deque              |    4.265 |    8 |         0 |
| chainyq.Deque_Ensure       |    3.460 |    0 |         0 |
| edwingeng.Deque            |    5.420 |    8 |         0 |
| gammazero.Deque            |    5.889 |   14 |         0 |
| gammazero.Deque_SetBaseCap |    3.885 |    8 |         0 |
| chainyq.List               |   38.640 |   24 |         1 |
| container.List             |   61.730 |   55 |         1 |

| PushFront                  | ns/op    | B/op | allocs/op |
|----------------------------|----------|------|-----------|
| chainyq.Deque              |    3.782 |    8 |         0 |
| edwingeng.Deque            |    4.273 |    8 |         0 |
| gammazero.Deque            |    4.732 |   17 |         0 |
| chainyq.List               |   33.800 |   24 |         1 |
| container.List             |   60.950 |   55 |         1 |

| Random churn (int)         | ns/op    | B/op | allocs/op |
|----------------------------|----------|------|-----------|
| chainyq.Deque              |    9.799 |    0 |         0 |
| edwingeng.Deque            |   10.350 |    0 |         0 |
| gammazero.Deque            |   10.880 |    0 |         0 |
| chainyq.List               |   18.560 |   11 |           |
| container.List             |   26.020 |   27 |           |

| Random churn (big struct)  | ns/op    | B/op | allocs/op |
|----------------------------|----------|------|-----------|
| chainyq.Deque              |    14.56 |    0 |         0 |
| edwingeng.Deque            |    15.61 |    0 |         0 |
| gammazero.Deque            |    15.06 |    0 |         0 |

| Random access (by index)   | ns/op    | B/op | allocs/op |
|----------------------------|----------|------|-----------|
| chainyq.Deque              |    28.67 |    0 |         0 |
| edwingeng.Deque            |  2808.00 |    0 |         0 |
| gammazero.Deque            |      DNF |    0 |         0 |

<p><sub>See full benchmark results in bench.txt, or run it yourself - code is available at `/deque/deque_benchmark_test.go`</sub></p>

```go
d := deque.FromSlice([]int{2, 4, 8, 16})
v, _ := d.Pop()  // 16, true
d.PushFront(1)
if v, ok := d.PopFront(); ok {  // v = 1
	fmt.Printf("popped from front: %v\n", v)
}
d.ToSlice() // [2 4 8]
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

The iterators have:
- Slicing: `Slice`/`PtrSlice`
- Collecting: `ToSlice`/`ToPtrSlice`/`ToChan`/`ToPtrChan`

Iterator API:
- Conversions: `Clone`/`Bidi`/`Rev`/`Seq`/`PtrSeq` + `RevSeq`/`RevPtrSeq` for bidi
- Resets: `Reset` + `ResetBack` for bidi
- Main: `HasNext`/`Next`/`NextPtr`/`Peek`/`PeekPtr` + `HasPrev`/`Prev`/`PrevPtr`/`PeekBack`/`PeekBackPtr` for bidi
- Actions on remaining items: `ForEach`/`ForEachPtr`
- Collecting remaining items: `ToChan`/`ToPtrChan`/`TakeSlice`/`TakePtrSlice`/`TakeWhile`/`TakeWhilePtr`
- Traversal: `Skip`/`SkipWhile`/`SkipWhilePtr` + `SkipBack` for bidi

See docs for the full API.

Synchronized version of the deque is available as `chainyq.deque.syncdeque.SyncDeque`.

## List[T any]
`list.List` is a doubly-linked list that can do all the operations you expect from a linked list and more. The list is not safe for concurrent use.
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

	l = list.FromSice[int]([]int{1, 2, 4, 8, 4, 2, 1})
	eq := func(a, b int) bool { return a == b }
	idx := l.IndexOf(2, eq)    // -> 1
	idx = l.LastIndexOf(2, eq) // -> 5
	idx = l.IndexOf(65535, eq) // -> -1
}
```

Three types of iterators - forward, reverse and bidirectional:
```go
import (
	"github.com/zelr0x/chainyq/list"
)

func main() {
	l = list.FromSice[int]([]int{1, 2, 4, 8, 4, 2, 1})

	l.Iter().Skip(2).ForEachPtr(func(v *int) bool {
		*v *= 2
		return true
	})
	l.String() // -> List[1 2 8 16 8 4 2]

	it := l.BidiIter()
	v, ok := v.Prev() // -> 0, false
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
	v, ok = it.Current() // -> 100, true
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


## Seq[T any]
`seq.Seq` is a lazy sequence that can be created from anything that generates values
i.e. `func(v T) bool`:
```go
import (
	"github.com/zelr0x/chainyq/seq"
)

func main() {
	s := seq.FromSlice([]string{"apple", "apricot", "banana", "cake"}).
	Filter(func(s string) bool {
		return s[0] != 'c'
	})
	m := seq.GroupBy(s, func(s string) string { return string(s[0])})
	m  // {"a": ["apple", "apricot"], "b": "banana"}
}
```
Operations that leave the sequence's element type unchanged are defined as methods on `seq.Seq`.
Operations that transform the element type are provided as functions in the `seq` package.

See docs for the full list of available operations.


`list.List[T]`'s iterators can be converted to sequences with the following methods:
```go
(it *Iter[T]) Seq() seq.Seq[T]
(it *Iter[T]) PtrSeq() seq.Seq[*T]
(it *RevIter[T]) Seq() seq.Seq[T]
(it *RevIter[T]) PtrSeq() seq.Seq[*T]
```
