# chainyq

<img src="docs/images/logo.png" alt="Project Logo" width="250"/>

Chainyq provides generic lists and queues with a rich, flexible API. Written in pure Go with zero dependencies.

## List
List is a doubly-linked list that can do all the operations you expect from a linked list and more:
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

Iterators:
```go
import (
    "github.com/zelr0x/chainyq/list"
)

func main() {
    l = list.FromSice[int]([]int{1, 2, 4, 8, 4, 2, 1})

    l.Iter().Skip(2).ForEachPtr(func(v *int) bool {
        return *v *= 2
    })
    l.String() // -> List[1, 2, 8, 16, 8, 4, 2]

    it := l.Iter()
    v, ok := v.Prev() // -> 0, false
    v, ok = it.Peek() // -> 1, true
    v, ok = it.Next() // -> 1, true
    ok = it.InsertBefore(55) // -> true
    v, ok = it.Prev() // -> 55, true
    v, ok = it.Remove() // -> 55, true
    v, ok = it.Next() // -> 1, true
    v, ok = it.PeekBack() // 0, false

    l.String() // -> List[1, 2, 8, 16, 8, 4, 2]

    v, ok = it.ResetBack().PrevPtr() // -> int* pointing to 2, ok
    *v += 98
    v, ok = it.Current() // -> 100, true
    ok = it.InsertAfter(42)
    l.String() // -> List[1, 2, 8, 16, 8, 4, 100, 42]

    it.Reset() // -> move to head
        SkipWhile(func(v int) bool {
            return v < 16
        }).  // it.Next() would return 16 here
        SkipBack(1). // now it.Next() would return 8
        DrainWhile(func(v int) bool {
            return v != 4 // skip over [8, 16, 8]
        })
    slice1 := l.ToSlice() // [1, 2, 4, 100, 42]
    slice2 := it.Reset().Skip(1).TakeSlice(3) // [2, 4, 100]
}
```
See more in the documentation.
