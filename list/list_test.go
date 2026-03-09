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
	"testing"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestNewAndIsEmpty(t *testing.T) {
	l := New[int]()
	AssertTrue(t, l.IsEmpty(), "new list must be empty")
	AssertEq(t, 0, l.Len(), "new list must have length 0")
}

func TestFromSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"empty slice", []int{}, []int{}},
		{"single element", []int{42}, []int{42}},
		{"multiple elements", []int{1, 2, 3}, []int{1, 2, 3}},
		{"large slice", SliceFromRangeExcl(t, 0, 150), SliceFromRangeExcl(t, 0, 150)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.input)
			got := l.ToSlice()
			AssertSliceEq(t, tt.want, got)
		})
	}
}

func TestAppendAndLen(t *testing.T) {
	var l *List[int]
	AssertEq(t, 0, Len(l))
	l = Append(l, 1)
	AssertEq(t, 1, Len(l))
}

func TestAddAndLen(t *testing.T) {
	tests := []struct {
		name    string
		want    []int
		wantLen int
	}{
		{"single add", []int{42}, 1},
		{"multiple adds", []int{1, 2, 3}, 3},
		{"no adds", []int{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New[int]()
			for _, v := range tt.want {
				l.Add(v)
			}
			AssertEq(t, tt.wantLen, l.Len())
			AssertSliceEq(t, tt.want, l.ToSlice())
		})
	}
}

func TestString(t *testing.T) {
	var l *List[int]
	AssertEq(t, "nil", l.String())
	l = New[int]()
	AssertEq(t, "List[]", l.String())
	l.Add(1).Add(2).Add(3)
	AssertEq(t, "List[1 2 3]", l.String())
}

func TestGoString(t *testing.T) {
	var l *List[int]
	AssertEq(t, "nil", l.String())
	l = New[int]()
	AssertEq(t, "List[int]{}", fmt.Sprintf("%#v", l))
	l.Add(1).Add(2).Add(3)
	AssertEq(t, "List[int]{1, 2, 3}", fmt.Sprintf("%#v", l))
}

func TestEquals(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name string
		a    *List[int]
		b    *List[int]
		want bool
	}{
		{"both nil", nil, nil, true},
		{"lhs nil rhs empty", nil, FromSlice([]int{}), false},
		{"lhs empty rhs nil", FromSlice([]int{}), nil, false},
		{"both empty", FromSlice([]int{}), FromSlice([]int{}), true},
		{"same elements", FromSlice([]int{1, 2, 3}), FromSlice([]int{1, 2, 3}), true},
		{"different lengths", FromSlice([]int{1, 2}), FromSlice([]int{1, 2, 3}), false},
		{"different elements", FromSlice([]int{1, 2, 3}), FromSlice([]int{1, 2, 4}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEq(t, tt.want, tt.a.Equals(tt.b, eq))
		})
	}
}

func TestFrontAndBack(t *testing.T) {
	tests := []struct {
		name        string
		values      []int
		wantFront   int
		wantFrontOK bool
		wantBack    int
		wantBackOK  bool
	}{
		{"empty list", []int{}, 0, false, 0, false},
		{"single element", []int{42}, 42, true, 42, true},
		{"multiple elements", []int{1, 2, 3}, 1, true, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got, ok := l.Front()
			AssertEq(t, tt.wantFront, got)
			AssertEq(t, tt.wantFrontOK, ok)

			got, ok = l.Back()
			AssertEq(t, tt.wantBack, got)
			AssertEq(t, tt.wantBackOK, ok)
		})
	}
}

func TestFrontPtrAndBackPtr(t *testing.T) {
	l := New[int]()

	got, ok := l.FrontPtr()
	AssertZeroFalse(t, got, ok)
	got, ok = l.BackPtr()
	AssertZeroFalse(t, got, ok)

	l.PushBack(10)
	l.PushBack(20)
	front, okFront := l.FrontPtr()
	AssertNotNil(t, front, "FrontPtr")
	AssertEqOk(t, 10, *front, okFront, "FrontPtr")
	back, okBack := l.BackPtr()
	AssertNotNil(t, front, "BackPtr")
	AssertEqOk(t, 20, *back, okBack, "BackPtr")
}

func TestPushFrontAndPushBack(t *testing.T) {
	l := New[int]()
	l.PushBack(2)
	l.PushFront(1)
	l.PushBack(3)
	front, _ := l.Front()
	back, _ := l.Back()
	AssertEq(t, 1, front)
	AssertEq(t, 3, back)
	AssertEq(t, 3, l.Len())
}

func TestPopFront(t *testing.T) {
	tests := []struct {
		name    string
		values  []int
		wantVal int
		wantOK  bool
		wantLen int
	}{
		{"empty list", []int{}, 0, false, 0},
		{"single element", []int{42}, 42, true, 0},
		{"multiple elements", []int{1, 2, 3}, 1, true, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got, ok := l.PopFront()
			AssertEq(t, tt.wantVal, got)
			AssertEq(t, tt.wantOK, ok)
			AssertEq(t, tt.wantLen, l.Len())
		})
	}
}

func TestPopBack(t *testing.T) {
	tests := []struct {
		name    string
		values  []int
		wantVal int
		wantOK  bool
		wantLen int
	}{
		{"empty list", []int{}, 0, false, 0},
		{"single element", []int{99}, 99, true, 0},
		{"multiple elements", []int{1, 2, 3}, 3, true, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got, ok := l.PopBack()
			AssertEq(t, tt.wantVal, got)
			AssertEq(t, tt.wantOK, ok)
			AssertEq(t, tt.wantLen, l.Len())
		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		idx        int
		val        int
		wantOK     bool
		wantResult []int
	}{
		{"insert at front", []int{2, 3}, 0, 1, true, []int{1, 2, 3}},
		{"insert in middle", []int{1, 3}, 1, 2, true, []int{1, 2, 3}},
		{"insert at back", []int{1, 2}, 2, 3, true, []int{1, 2, 3}},
		{"insert out of range negative", []int{1, 2}, -1, 9, false, []int{1, 2}},
		{"insert out of range too large", []int{1, 2}, 5, 9, false, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.initial)
			ok := l.Insert(tt.idx, tt.val)
			AssertEq(t, tt.wantOK, ok)
			got := l.ToSlice()
			AssertSliceEq(t, tt.wantResult, got)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		idx        int
		wantVal    int
		wantOK     bool
		wantResult []int
	}{
		{"remove front", []int{1, 2, 3}, 0, 1, true, []int{2, 3}},
		{"remove middle", []int{1, 2, 3}, 1, 2, true, []int{1, 3}},
		{"remove back", []int{1, 2, 3}, 2, 3, true, []int{1, 2}},
		{"remove out of range negative", []int{1, 2}, -1, 0, false, []int{1, 2}},
		{"remove out of range too large", []int{1, 2}, 5, 0, false, []int{1, 2}},
		{"remove from empty", []int{}, 0, 0, false, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.initial)
			val, ok := l.Remove(tt.idx)
			AssertEq(t, tt.wantVal, val)
			AssertEq(t, tt.wantOK, ok)
			got := l.ToSlice()
			AssertSliceEq(t, tt.wantResult, got)
		})
	}
}

func TestInsertAndRemoveLarge(t *testing.T) {
	const size = 150
	l := listFromRangeExcl(t, 0, size)

	// Insert in the middle
	AssertTrue(t, l.Insert(size/2, 999), "Insert at middle")
	AssertEq(t, size+1, l.Len(), "length changed after insert")

	val, ok := l.Get(size / 2)
	AssertEqOk(t, 999, val, ok, "get inserted in middle returns expected value")

	// Remove from the middle
	val, ok = l.Remove(size/2)
	AssertEqOk(t, 999, val, ok, "Remove at middle")
	AssertEq(t, size, l.Len(), "size after Remove")

	val, ok = l.Get(size / 2)
	AssertTrue(t, ok)
	AssertNotEq(t, 999, val, "value 999 should have been removed")

	val, ok = l.Remove(0)
	AssertEqOk(t, 0, val, ok, "Remove first item")
	AssertEq(t, size-1, l.Len(), "Len after one item removed")
	val, ok = l.Remove(l.Len()-1)
	AssertEqOk(t, 149, val, ok, "Remove last item")
	AssertEq(t, size-2, l.Len(), "Len after two items removed")
}

func TestIndexOf(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name    string
		values  []int
		target  int
		wantIdx int
	}{
		{"empty list", []int{}, 5, -1},
		{"not found", []int{1, 2, 3}, 9, -1},
		{"found at front", []int{1, 2, 3}, 1, 0},
		{"found in middle", []int{1, 2, 3}, 2, 1},
		{"found at back", []int{1, 2, 3}, 3, 2},
		{"duplicate elements", []int{1, 2, 2, 3}, 2, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got := l.IndexOf(tt.target, eq)
			AssertEq(t, tt.wantIdx, got, fmt.Sprintf("IndexOf(%d)", tt.target))
		})
	}
}

func TestLastIndexOf(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name    string
		values  []int
		target  int
		wantIdx int
	}{
		{"empty list", []int{}, 5, -1},
		{"not found", []int{1, 2, 3}, 9, -1},
		{"found at front", []int{1, 2, 3}, 1, 0},
		{"found in middle", []int{1, 2, 3}, 2, 1},
		{"found at back", []int{1, 2, 3}, 3, 2},
		{"duplicate elements", []int{1, 2, 2, 3}, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got := l.LastIndexOf(tt.target, eq)
			AssertEq(t, tt.wantIdx, got, fmt.Sprintf("LastIndexOf(%d)", tt.target))
		})
	}
}

func TestContains(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name   string
		values []int
		target int
		want   bool
	}{
		{"empty list", []int{}, 5, false},
		{"not found", []int{1, 2, 3}, 9, false},
		{"found at front", []int{1, 2, 3}, 1, true},
		{"found in middle", []int{1, 2, 3}, 2, true},
		{"found at back", []int{1, 2, 3}, 3, true},
		{"duplicate elements", []int{1, 2, 2, 3}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			got := l.Contains(tt.target, eq)
			AssertEq(t, tt.want, got, fmt.Sprintf("Contains(%d)", tt.target))
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		values  []int
		idx     int
		wantVal int
		wantOK  bool
	}{
		{"empty list", []int{}, 0, 0, false},
		{"negative index", []int{1, 2, 3}, -1, 0, false},
		{"index too large", []int{1, 2, 3}, 5, 0, false},
		{"get front", []int{1, 2, 3}, 0, 1, true},
		{"get middle", []int{1, 2, 3}, 1, 2, true},
		{"get back", []int{1, 2, 3}, 2, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.values)
			gotVal, gotOK := l.Get(tt.idx)
			AssertCommaOk(t, tt.wantVal, tt.wantOK, gotVal, gotOK, fmt.Sprintf("Get(%d)", tt.idx))
		})
	}
}

func TestForEach(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	var visited []int
	l.ForEach(func(v int) bool {
		visited = append(visited, v)
		return true
	})
	want := []int{1, 2, 3}
	AssertSliceEq(t, want, visited)
	// Early stop
	visited = nil
	l.ForEach(func(v int) bool {
		visited = append(visited, v)
		return v < 2
	})
	AssertEq(t, 2, len(visited), "expected early stop after 2 items")
}

func TestForEachPtr(t *testing.T) {
	l := New[int]().Add(10).Add(20)
	l.ForEachPtr(func(p *int) bool {
		AssertNotNil(t, p)
		*p *= 2
		return true
	})
	AssertSliceEq(t, []int{20, 40}, l.ToSlice())
}

func TestRemoveIf(t *testing.T) {
	l := listFromRangeIncl(t, 1, 5)
	removed := l.RemoveIf(func(v int) bool { return v%2 == 0 })
	AssertEq(t, 2, removed, "expected 2 items removed")
	got := l.ToSlice()
	AssertSliceEq(t, []int{1, 3, 5}, got)
}

func TestConcat(t *testing.T) {
	l1 := FromSlice([]int{1, 2})
	l2 := FromSlice([]int{3, 4})
	l1.Concat(l2)
	got := l1.ToSlice()
	want := []int{1, 2, 3, 4}
	AssertSliceEq(t, want, got)
	AssertTrue(t, l2.IsEmpty(), "expected second list to be empty after Concat")
}

func TestClear(t *testing.T) {
	l := listFromRangeIncl(t, 1, 20)
	l.Clear()
	AssertTrue(t, l.IsEmpty(), "list should be empty after Clear")
	AssertZero(t, l.Len(), "list Len should be 0 after Clear")
}

func TestNewIterNotNil(t *testing.T) {
	l := New[int]()
	AssertNotNil(t, l.Iter())
	l.PushBack(1)
	AssertNotNil(t, l.Iter())
}

func TestNewRevIterNotNil(t *testing.T) {
	l := New[int]()
	AssertNotNil(t, l.RevIter())
	l.PushBack(1)
	AssertNotNil(t, l.RevIter())
}

func TestNewBidiIterNotNil(t *testing.T) {
	l := New[int]()
	AssertNotNil(t, l.BidiIter())
	l.PushBack(1)
	AssertNotNil(t, l.BidiIter())
}

func TestSliceAndPtrSlice(t *testing.T) {
	lEven := listFromRangeIncl(t, 1, 6)
	lOdd := listFromRangeIncl(t, 1, 5)
	tests := []struct {
		name     string
		list     *List[int]
		start    int
		end      int
		wantVals []int
	}{
		{"empty slice from empty list", New[int](), 0, 3, []int{}},
		{"full slice even length", lEven, 0, 6, []int{1, 2, 3, 4, 5, 6}},
		{"full slice odd length", lOdd, 0, 5, []int{1, 2, 3, 4, 5}},
		{"partial slice middle", lEven, 2, 5, []int{3, 4, 5}},
		{"slice with start=0 end=0", lEven, 0, 0, []int{}},
		{"slice with start=end", lEven, 3, 3, []int{}},
		{"slice with start>end", lEven, 4, 2, []int{}},
		{"slice with negative start", lEven, -2, 3, []int{}},
		{"slice with negative end", lEven, 0, -1, []int{}},
		{"slice with end beyond length", lOdd, 0, 10, []int{1, 2, 3, 4, 5}},
		{"slice with start beyond length", lOdd, 10, 15, []int{}},
		{"slice covering tail", lEven, 4, 10, []int{5, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.list.Slice(tt.start, tt.end)
			AssertSliceEq(t, tt.wantVals, got, "Slice")
			ptrGot := tt.list.PtrSlice(tt.start, tt.end)
			AssertPtrSliceEq(t, tt.wantVals, ptrGot, "PtrSlice")
		})
	}
}

func TestToSliceAndToPtrSlice(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	got := l.ToSlice()
	want := []int{1, 2, 3}
	AssertSliceEq(t, want, got)
	ptrGot := l.ToPtrSlice()
	AssertPtrSliceEq(t, want, ptrGot)
}

func TestToChanAndToPtrChan(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)

	ch := l.ToChan(0)
	var got []int
	for v := range ch {
		got = append(got, v)
	}
	want := []int{1, 2, 3}
	AssertSliceEq(t, want, got, "ToChan")

	ptrCh := l.ToPtrChan(0)
	var ptrGot []int
	for p := range ptrCh {
		if p == nil {
			t.Errorf("unexpected nil pointer from ToPtrChan")
			continue
		}
		ptrGot = append(ptrGot, *p)
	}
	AssertSliceEq(t, want, ptrGot, "ToPtrChan")
}

func TestBidiIterDoesNotAllowToGoPastSentinel(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	it := l.BidiIter()
	it.Reset()
	val, ok := it.Prev()
	AssertZeroFalse(t, val, ok, "Prev after Reset should be false")
	it.ResetBack()
	val, ok = it.Next()
	AssertZeroFalse(t, val, ok, "Next after ResetBack should be false")
}

func TestBidiIterCloneResetAndCurrent(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	it := l.BidiIter()
	AssertNotNil(t, it)
	val, ok := it.Current()
	AssertZeroFalse(t, val, ok, "Current should be false before Next")
	ptr, ok := it.CurrentPtr()
	AssertZeroFalse(t, ptr, ok, "CurrentPtr should be false before Next")

	// Advance once
	it.Next()
	val, ok = it.Current()
	AssertEqOk(t, 1, val, ok, "Current after next")
	ptr, ok = it.CurrentPtr()
	AssertNotNil(t, ptr, "CurrentPtr after next not nil")
	AssertEqOk(t, 1, *ptr, ok, "CurrentPtr after next")

	// Clone should point to same position
	clone := it.Clone()
	valClone, okClone := clone.Current()
	AssertEqOk(t, 1, valClone, okClone, "Current after Clone")

	// Reset should go back to head
	it.Reset().Next()
	val, ok = it.Current()
	AssertEqOk(t, 1, val, ok, "Current after Reset Next")

	// ResetBack should go to tail
	val, ok = it.ResetBack().Current()
	AssertZeroFalse(t, val, ok, "Current after ResetBack")

	it.Prev() // step back once to get to last item
	it.Next() // step forward once - should not go to tail
	val, ok = it.Current()
	AssertEqOk(t, 3, val, ok, "Current after ResetBack Prev Next")
}

func TestBidiIterEmptyList(t *testing.T) {
	l := New[int]()
	it := l.BidiIter()
	AssertNotNil(t, it)
	val, ok := it.Current()
	AssertZeroFalse(t, val, ok, "Current on empty list")
	ptr, ok := it.CurrentPtr()
	AssertZeroFalse(t, ptr, ok, "CurrentPtr on empty list")

	// Reset should still be safe
	it.Reset()
	val, ok = it.Current()
	AssertZeroFalse(t, val, ok, "Current after Reset on empty list")

	// ResetBack should also be safe
	it.ResetBack()
	val, ok = it.Current()
	AssertZeroFalse(t, val, ok, "Current after ResetBack on empty list")
}

func TestBidiIterNavigation(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	it := l.BidiIter()
	AssertTrue(t, it.HasNext(), "HasNext at start should be true")
	AssertFalse(t, it.HasPrev(), "HasPrev at start should be false")

	val, ok := it.Peek()
	AssertEqOk(t, 1, val, ok, "Peek should show first element without advancing")
	val, ok = it.Next()
	AssertEqOk(t, 1, val, ok, "Next should advance into first element")

	AssertFalse(t, it.HasPrev(), "HasPrev at first non-sentinel")

	val, ok = it.Next()
	AssertEqOk(t, 2, val, ok, "Next should advance further")

	AssertTrue(t, it.HasPrev(), "HasPrev at second element")

	val, ok = it.Prev()
	AssertEqOk(t, 1, val, ok, "Prev should step back")
	val, ok = it.Peek()
	AssertEqOk(t, 2, val, ok, "Peek after Prev should show second")
}

func TestBidiIterNavigationEmptyList(t *testing.T) {
	l := New[int]()
	it := l.BidiIter()

	AssertFalse(t, it.HasNext(), "HasNext on empty list")
	AssertFalse(t, it.HasPrev(), "HasPrev on empty list")

	val, ok := it.Next()
	AssertZeroFalse(t, val, ok, "Next on empty list")
	ptr, ok := it.NextPtr()
	AssertZeroFalse(t, ptr, ok, "NextPtr on empty list")

	val, ok = it.Prev()
	AssertZeroFalse(t, val, ok, "Prev on empty list")
	ptr, ok = it.PrevPtr()
	AssertZeroFalse(t, ptr, ok, "PrevPtr on empty list")

	val, ok = it.Peek()
	AssertZeroFalse(t, val, ok, "Peek on empty list")
	ptr, ok = it.PeekPtr()
	AssertZeroFalse(t, ptr, ok, "PeekPtr on empty list")

	val, ok = it.PeekBack()
	AssertZeroFalse(t, val, ok, "PeekBack on empty list")
	ptr, ok = it.PeekBackPtr()
	AssertZeroFalse(t, ptr, ok, "PeekBackPtr on empty list")
}

func TestBidiIterRemoveAndInsert(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	it := l.BidiIter()

	val, ok := it.Remove()
	AssertZeroFalse(t, val, ok, "Remove at head sentinel fails")
	AssertFalse(t, it.InsertBefore(99), "InsertBefore at head sentinel fails")

	it.Next()
	AssertTrue(t, it.InsertBefore(77), "InsertBefore at first item should succeed")
	AssertSliceEq(t, []int{77, 1, 2, 3}, l.ToSlice(), "InsertBefore at first item")

	AssertTrue(t, it.InsertAfter(88), "InsertAfter at first item should succeed")
	AssertSliceEq(t, []int{77, 1, 88, 2, 3}, l.ToSlice(), "InsertAfter at first item")

	it.Next()
	val, ok = it.Remove()
	AssertEqOk(t, 88, val, ok, "Remove at middle")
	AssertSliceEq(t, []int{77, 1, 2, 3}, l.ToSlice(), "Remove at middle")

	it.ResetBack()
	val, ok = it.Remove()
	AssertFalse(t, ok, "Remove at tail sentinel should fail")
	AssertFalse(t, it.InsertAfter(99), "InsertAfter at tail sentinel should fail")

	it.Prev()
	AssertTrue(t, it.InsertAfter(66), "InsertAfter at at last item should succeed")
	AssertSliceEq(t, []int{77, 1, 2, 3, 66}, l.ToSlice(), "InsertAfter at last item")
}

func TestIterForEachAndChannels(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	slice := []int{1, 2, 3}

	// ForEach should visit all items starting from current position
	it := l.Iter()
	var got []int
	it.ForEach(func(v int) bool {
		got = append(got, v)
		return true
	})
	AssertSliceEq(t, slice, got, "ForEach")

	// ForEachPtr should visit all items with pointers
	it = l.Iter()
	var ptrGot []int
	it.ForEachPtr(func(p *int) bool {
		ptrGot = append(ptrGot, *p)
		return true
	})
	AssertSliceEq(t, slice, ptrGot, "ForEachPtr")

	// ForEach can stop early if callback returns false
	it = l.Iter()
	got = nil
	it.ForEach(func(v int) bool {
		got = append(got, v)
		return v < 2 // stop after 2
	})
	var want []int
	for _, v := range slice {
		if v > 2 {
			break
		}
		want = append(want, v)
	}
	AssertSliceEq(t, want, got, "ForEach early stop")

	// ToChan should produce all remaining items
	it = l.Iter()
	ch := it.ToChan(0) // unbuffered
	var chanGot []int
	for v := range ch {
		chanGot = append(chanGot, v)
	}
	AssertSliceEq(t, slice, chanGot, "ToChan")

	// ToPtrChan should produce pointers
	it = l.Iter()
	ptrCh := it.ToPtrChan(2) // buffered
	var chanPtrGot []int
	for p := range ptrCh {
		chanPtrGot = append(chanPtrGot, *p)
	}
	AssertSliceEq(t, slice, chanPtrGot, "ToChanPtr")
}

func TestRevIterForEachAndChannels(t *testing.T) {
	l := listFromRangeIncl(t, 1, 3)
	revSlice := ReversedSlice(l.ToSlice())

	// ForEach should visit all items starting from current position
	it := l.RevIter()
	var got []int
	it.ForEach(func(v int) bool {
		got = append(got, v)
		return true
	})
	AssertSliceEq(t, revSlice, got, "ForEach")

	// ForEachPtr should visit all items with pointers
	it = l.RevIter()
	var ptrGot []int
	it.ForEachPtr(func(p *int) bool {
		ptrGot = append(ptrGot, *p)
		return true
	})
	AssertSliceEq(t, revSlice, ptrGot, "ForEachPtr")

	// ForEach can stop early if callback returns false
	it = l.RevIter()
	got = nil
	it.ForEach(func(v int) bool {
		got = append(got, v)
		return v > 2 // stop after 2
	})
	var want []int
	for _, v := range revSlice {
		if v < 2 {
			break
		}
		want = append(want, v)
	}
	AssertSliceEq(t, want, got, "ForEach early stop")

	it = l.RevIter()
	ch := it.ToChan(0) // unbuffered
	var chanGot []int
	for v := range ch {
		chanGot = append(chanGot, v)
	}
	AssertSliceEq(t, revSlice, chanGot, "ToChan")

	it = l.RevIter()
	ptrCh := it.ToPtrChan(2) // buffered
	var chanPtrGot []int
	for p := range ptrCh {
		chanPtrGot = append(chanPtrGot, *p)
	}
	AssertSliceEq(t, revSlice, chanPtrGot, "ToPtrChan")
}

func TestIterForEachAndChannelsEmpty(t *testing.T) {
	l := New[int]()
	tests := []struct {
		name string
		it   interface {
			ForEach(func(int) bool)
			ToChan(cap int) <-chan int
			ToPtrChan(cap int) <-chan *int
		}
	}{
		{"Iter", l.Iter()},
		{"BidiIter", l.BidiIter()},
		{"RevIter", l.RevIter()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			tt.it.ForEach(func(v int) bool {
				called = true
				return true
			})
			AssertFalse(t, called, "ForEach on empty list should not call the arg func")

			ch := tt.it.ToChan(1)
			if v, ok := <-ch; ok {
				t.Errorf("%s: ToChan on empty list yielded %d", tt.name, v)
			}
			ptrCh := tt.it.ToPtrChan(1)
			if p, ok := <-ptrCh; ok {
				t.Errorf("%s: ToPtrChan on empty list yielded %v", tt.name, p)
			}
		})
	}
}

func TestIterTakeAndSkip(t *testing.T) {
	n := 5
	items := SliceFromRangeIncl(t, 1, n)
	revItems := ReversedSlice(items)
	l := FromSlice(items)
	tests := []struct {
		name  string
		it    ListIterator[int]
		items []int
	}{
		{"Iter", l.Iter(), items},
		{"BidiIter", l.BidiIter(), items},
		{"RevIter", l.RevIter(), revItems},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := tt.it
			k := n / 2
			got := it.TakeSlice(k)
			want := tt.items[:k]
			AssertSliceEq(t, want, got, "TakeSlice")
			ptrGot := it.TakePtrSlice(n - k)
			want = tt.items[k:]
			AssertPtrSliceEq(t, want, ptrGot, "TakePtrSlice")
		})
	}
}

func TestIterTakeWhileAndSkip(t *testing.T) {
	l := listFromRangeIncl(t, 1, 5)

	// TakeWhile should stop at first non-match
	it := l.BidiIter()
	got := it.TakeWhile(func(v int) bool { return v < 3 })
	AssertSliceEq(t, []int{1, 2}, got, "TakeWhile")

	// TakeWhilePtr should stop at first non-match
	it = l.BidiIter()
	ptrs := it.TakeWhilePtr(func(p *int) bool { return *p < 4 })
	AssertPtrSliceEq(t, []int{1, 2, 3}, ptrs, "TakeWhilePtr")

	// Skip should advance by n
	it = l.BidiIter()
	it.Skip(2)
	val, ok := it.Next()
	AssertEqOk(t, 3, val, ok, "Skip(2) then Next")

	// SkipBack should step back by n
	it.ResetBack()
	it.SkipBack(2)
	val, ok = it.Prev()
	AssertEqOk(t, 3, val, ok, "SkipBack(2) then Prev")

	// SkipWhile should advance until predicate fails
	it = l.BidiIter()
	it.SkipWhile(func(v int) bool { return v < 4 })
	val, ok = it.Next()
	AssertEqOk(t, 4, val, ok, "SkipWhile then Next")

	// SkipWhilePtr should advance until predicate fails
	it = l.BidiIter()
	it.SkipWhilePtr(func(p *int) bool { return *p < 5 })
	val, ok = it.Next()
	AssertEqOk(t, 5, val, ok, "SkipWhilePtr then Next")

	// DrainWhile should remove items until predicate fails
	it = l.BidiIter()
	it.DrainWhile(func(v int) bool { return v < 3 })
	got = l.ToSlice()
	AssertSliceEq(t, []int{3, 4, 5}, got, "DrainWhile modified list as expected")

	// DrainWhilePtr should remove items until predicate fails
	l = listFromRangeIncl(t, 1, 5)
	it = l.BidiIter()
	it.DrainWhilePtr(func(p *int) bool { return *p < 4 })
	got = l.ToSlice()
	AssertSliceEq(t, []int{4, 5}, got, "DrainWhilePtr modified list as expected")
}

func TestIterTakeSkipEmpty(t *testing.T) {
	l := New[int]()
	it := l.BidiIter()

	AssertSliceEq(t, []int{}, it.TakeSlice(3), "TakeSlice(3)")
	AssertSliceEq(t, []*int{}, it.TakePtrSlice(3), "TakePtrSlice(3)")
	AssertSliceEq(t, []int{}, it.TakeWhile(func(v int) bool { return true }), "TakeWhile")
	AssertSliceEq(t, []*int{}, it.TakeWhilePtr(func(v *int) bool { return true }), "TakeWhilePtr")

	// Skip/SkipBack/SkipWhile/DrainWhile should be safe no-ops
	it.
		Skip(5).
		SkipBack(5).
		SkipWhile(func(v int) bool {
			return true
		}).SkipWhilePtr(func(p *int) bool {
		return true
	}).DrainWhile(func(v int) bool {
		return true
	}).DrainWhilePtr(func(p *int) bool {
		return true
	})
	got, ok := it.Current()
	AssertZeroFalse(t, got, ok,
		"Skip SkipBack SkipWhile DrainWhile DrainWhilePtr on empty")
}

func TestDocExample1(t *testing.T) {
	l := New[int]()
	l.Add(2).Add(4)
	l.PushBack(8)
	l.PushFront(1)
	AssertEq(t, "List[1 2 4 8]", l.String(), "String")

	l = FromSlice([]int{1, 2, 4, 8, 4, 2, 1})
	eq := func(a, b int) bool { return a == b }
	AssertEq(t, 1, l.IndexOf(2, eq), "IndexOf(2)")
	AssertEq(t, 5, l.LastIndexOf(2, eq), "LastIndexof(2)")
	AssertEq(t, -1, l.IndexOf(65535, eq), "IndexOf(65535)")

	l.Iter().Skip(2).ForEachPtr(func(v *int) bool {
		*v *= 2
		return true
	})
	want := []int{1, 2, 8, 16, 8, 4, 2}
	AssertSliceEq(t, want, l.ToSlice(), "List items doubled with ForEachPtr")

	it := l.BidiIter()
	v, ok := it.Prev()
	AssertZeroFalse(t, v, ok, "Prev at head")
	v, ok = it.Peek()
	AssertEqOk(t, 1, v, ok, "Peek at head non-empty")
	v, ok = it.Peek()
	AssertEqOk(t, 1, v, ok, "Peek at head non-empty, consecutive")
	v, ok = it.Next()
	AssertEqOk(t, 1, v, ok, "Next at head non-empty")
	ok = it.InsertBefore(55)
	AssertEq(t, true, ok, "InsertBefore first item")
	v, ok = it.Prev()
	AssertEqOk(t, 55, v, ok, "Prev at first item after InsertBefore")
	v, ok = it.Remove()
	AssertEqOk(t, 55, v, ok, "Remove at value 55")
	v, ok = it.Next()
	AssertEqOk(t, 1, v, ok, "Next")
	v, ok = it.PeekBack()
	AssertZeroFalse(t, v, ok, "PeekBack at first item")
	AssertSliceEq(t, want, l.ToSlice(), "ops cancelled each other")

	p, ok := it.ResetBack().PrevPtr()
	AssertNotNil(t, p)
	AssertEqOk(t, 2, *p, ok, "ResetBack then PrevPtr")
	*p += 98
	it.InsertAfter(42)
	AssertSliceEq(t, []int{1, 2, 8, 16, 8, 4, 100, 42}, l.ToSlice(), "ResetBack PrevPtr modified")

	it.Reset().
		SkipWhile(func(v int) bool {
			return v < 16
		}).
		SkipBack(1).
		DrainWhile(func(v int) bool {
			return v != 4
		})
	AssertSliceEq(t, []int{1, 2, 4, 100, 42}, l.ToSlice(), "State after piped modifications")
	s := it.Reset().Skip(1).TakeSlice(3)
	AssertSliceEq(t, []int{2, 4, 100}, s, "State after Reset-Skip-TakeSlice")
}

// ----- Helpers -----
func listFromRangeIncl(t *testing.T, from, toIncl int) *List[int] {
	t.Helper()
	ValidRangeOrDie(t, from, toIncl)
	res := New[int]()
	for i := from; i <= toIncl; i++ {
		res.PushBack(i)
	}
	return res
}

func listFromRangeExcl(t *testing.T, from, to int) *List[int] {
	t.Helper()
	ValidRangeOrDie(t, from, to)
	res := New[int]()
	for i := from; i < to; i++ {
		res.PushBack(i)
	}
	return res
}
