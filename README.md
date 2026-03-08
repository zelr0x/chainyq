# chainyq

<img src="docs/images/logo.png" alt="Project Logo" width="250"/>

Chainyq provides generic lists and queues with a rich, flexible API. Written in pure Go with zero dependencies.

## List
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

## Seq
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
