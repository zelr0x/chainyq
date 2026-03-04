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
	"slices"
	"testing"
)

func TestNewAndIsEmpty(t *testing.T) {
	l := New[int]()
	if !l.IsEmpty() {
		t.Errorf("expected new list to be empty")
	}
	if got := l.Len(); got != 0 {
		t.Errorf("expected length 0, got %d", got)
	}
}

func TestFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{"empty slice", []int{}, []int{}},
		{"single element", []int{42}, []int{42}},
		{"multiple elements", []int{1, 2, 3}, []int{1, 2, 3}},
		{"large slice", func() []int {
			s := make([]int, 150)
			for i := 0; i < 150; i++ {
				s[i] = i
			}
			return s
		}(), func() []int {
			s := make([]int, 150)
			for i := 0; i < 150; i++ {
				s[i] = i
			}
			return s
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice(tt.input)
			got := l.ToSlice()
			if len(got) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(got))
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("at index %d, got %d, want %d", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestAppendAndLen(t *testing.T) {
	var l *List[int]
	if got := Len(l); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
	l = Append(l, 1)
	if got := Len(l); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestAddAndLen(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		expected int
	}{
		{"single add", []int{42}, 1},
		{"multiple adds", []int{1, 2, 3}, 3},
		{"no adds", []int{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := FromSlice[int](tt.values)
			if got := l.Len(); got != tt.expected {
				t.Errorf("expected length %d, got %d", tt.expected, got)
			}
			if (tt.expected == 0) != l.IsEmpty() {
				t.Errorf("IsEmpty mismatch: expected %v", tt.expected == 0)
			}
		})
	}
}

func TestString(t *testing.T) {
	l := New[int]()
	if got := l.String(); got != "List[]" {
		t.Errorf("expected empty list 'List[]', got %q", got)
	}
	l.Add(1).Add(2).Add(3)
	got := l.String()
	if got != "List[1, 2, 3]" {
		t.Errorf("unexpected string representation: %q", got)
	}
}

func TestEquals(t *testing.T) {
	eq := func(a, b int) bool { return a == b }

	tests := []struct {
		name     string
		a        []int
		b        []int
		expected bool
	}{
		{"both empty", []int{}, []int{}, true},
		{"same elements", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different lengths", []int{1, 2}, []int{1, 2, 3}, false},
		{"different elements", []int{1, 2, 3}, []int{1, 2, 4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l1 := FromSlice[int](tt.a)
			l2 := FromSlice[int](tt.b)
			if got := l1.Equals(l2, eq); got != tt.expected {
				t.Errorf("Equals mismatch: expected %v, got %v", tt.expected, got)
			}
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
			l := FromSlice[int](tt.values)
			if got, ok := l.Front(); got != tt.wantFront || ok != tt.wantFrontOK {
				t.Errorf("Front() = (%d,%v), want (%d,%v)", got, ok, tt.wantFront, tt.wantFrontOK)
			}
			if got, ok := l.Back(); got != tt.wantBack || ok != tt.wantBackOK {
				t.Errorf("Back() = (%d,%v), want (%d,%v)", got, ok, tt.wantBack, tt.wantBackOK)
			}
		})
	}
}

func TestFrontPtrAndBackPtr(t *testing.T) {
	l := New[int]()
	if got, ok := l.FrontPtr(); got != nil || ok {
		t.Errorf("FrontPtr on empty list = (%v,%v), want (nil,false)", got, ok)
	}
	if got, ok := l.BackPtr(); got != nil || ok {
		t.Errorf("BackPtr on empty list = (%v,%v), want (nil,false)", got, ok)
	}
	l.PushBack(10)
	l.PushBack(20)
	front, okFront := l.FrontPtr()
	back, okBack := l.BackPtr()
	if !okFront || front == nil || *front != 10 {
		t.Errorf("FrontPtr = (%v,%v), want (10,true)", front, okFront)
	}
	if !okBack || back == nil || *back != 20 {
		t.Errorf("BackPtr = (%v,%v), want (20,true)", back, okBack)
	}
}

func TestPushFrontAndPushBack(t *testing.T) {
	l := New[int]()
	l.PushBack(2)
	l.PushFront(1)
	l.PushBack(3)
	front, _ := l.Front()
	back, _ := l.Back()
	if front != 1 {
		t.Errorf("expected front=1, got %d", front)
	}
	if back != 3 {
		t.Errorf("expected back=3, got %d", back)
	}
	if got := l.Len(); got != 3 {
		t.Errorf("expected length 3, got %d", got)
	}
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
			l := FromSlice[int](tt.values)
			got, ok := l.PopFront()
			if got != tt.wantVal || ok != tt.wantOK {
				t.Errorf("PopFront() = (%d,%v), want (%d,%v)", got, ok, tt.wantVal, tt.wantOK)
			}
			if l.Len() != tt.wantLen {
				t.Errorf("after PopFront, Len() = %d, want %d", l.Len(), tt.wantLen)
			}
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
			l := FromSlice[int](tt.values)
			got, ok := l.PopBack()
			if got != tt.wantVal || ok != tt.wantOK {
				t.Errorf("PopBack() = (%d,%v), want (%d,%v)", got, ok, tt.wantVal, tt.wantOK)
			}
			if l.Len() != tt.wantLen {
				t.Errorf("after PopBack, Len() = %d, want %d", l.Len(), tt.wantLen)
			}
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
			l := New[int]()
			for _, v := range tt.initial {
				l.PushBack(v)
			}
			ok := l.Insert(tt.idx, tt.val)
			if ok != tt.wantOK {
				t.Errorf("Insert(%d,%d) ok=%v, want %v", tt.idx, tt.val, ok, tt.wantOK)
			}
			got := l.ToSlice()
			if len(got) != len(tt.wantResult) {
				t.Fatalf("expected slice length %d, got %d", len(tt.wantResult), len(got))
			}
			for i := range got {
				if got[i] != tt.wantResult[i] {
					t.Errorf("at index %d, got %d, want %d", i, got[i], tt.wantResult[i])
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		idx        int
		wantOK     bool
		wantResult []int
	}{
		{"remove front", []int{1, 2, 3}, 0, true, []int{2, 3}},
		{"remove middle", []int{1, 2, 3}, 1, true, []int{1, 3}},
		{"remove back", []int{1, 2, 3}, 2, true, []int{1, 2}},
		{"remove out of range negative", []int{1, 2}, -1, false, []int{1, 2}},
		{"remove out of range too large", []int{1, 2}, 5, false, []int{1, 2}},
		{"remove from empty", []int{}, 0, false, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New[int]()
			for _, v := range tt.initial {
				l.PushBack(v)
			}
			ok := l.Remove(tt.idx)
			if ok != tt.wantOK {
				t.Errorf("Remove(%d) ok=%v, want %v", tt.idx, ok, tt.wantOK)
			}
			got := l.ToSlice()
			if len(got) != len(tt.wantResult) {
				t.Fatalf("expected slice length %d, got %d", len(tt.wantResult), len(got))
			}
			for i := range got {
				if got[i] != tt.wantResult[i] {
					t.Errorf("at index %d, got %d, want %d", i, got[i], tt.wantResult[i])
				}
			}
		})
	}
}

func TestInsertAndRemoveLarge(t *testing.T) {
	const size = 150
	l := New[int]()
	for i := 0; i < size; i++ {
		l.PushBack(i)
	}
	// Insert in the middle
	if !l.Insert(size/2, 999) {
		t.Errorf("Insert at middle failed")
	}
	if l.Len() != size+1 {
		t.Errorf("expected length %d, got %d", size+1, l.Len())
	}
	if val, _ := l.Get(size / 2); val != 999 {
		t.Errorf("expected inserted value 999 at index %d, got %d", size/2, val)
	}
	// Remove from the middle
	if !l.Remove(size / 2) {
		t.Errorf("Remove at middle failed")
	}
	if l.Len() != size {
		t.Errorf("expected length %d after remove, got %d", size, l.Len())
	}
	if val, _ := l.Get(size / 2); val == 999 {
		t.Errorf("value 999 should have been removed")
	}
	// Remove front and back
	if !l.Remove(0) {
		t.Errorf("Remove front failed")
	}
	if !l.Remove(l.Len() - 1) {
		t.Errorf("Remove back failed")
	}
	if l.Len() != size-2 {
		t.Errorf("expected length %d after front/back remove, got %d", size-2, l.Len())
	}
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
			l := New[int]()
			for _, v := range tt.values {
				l.PushBack(v)
			}
			got := l.IndexOf(tt.target, eq)
			if got != tt.wantIdx {
				t.Errorf("IndexOf(%d) = %d, want %d", tt.target, got, tt.wantIdx)
			}
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
			l := New[int]()
			for _, v := range tt.values {
				l.PushBack(v)
			}
			got := l.LastIndexOf(tt.target, eq)
			if got != tt.wantIdx {
				t.Errorf("LastIndexOf(%d) = %d, want %d", tt.target, got, tt.wantIdx)
			}
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
			l := FromSlice[int](tt.values)
			got := l.Contains(tt.target, eq)
			if got != tt.want {
				t.Errorf("Contains(%d) = %v, want %v", tt.target, got, tt.want)
			}
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
			l := FromSlice[int](tt.values)
			gotVal, gotOK := l.Get(tt.idx)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("Get(%d) = (%d,%v), want (%d,%v)", tt.idx, gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestForEach(t *testing.T) {
	l := New[int]()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	var visited []int
	l.ForEach(func(v int) bool {
		visited = append(visited, v)
		return true
	})
	want := []int{1, 2, 3}
	if len(visited) != len(want) {
		t.Fatalf("expected %d elements, got %d", len(want), len(visited))
	}
	for i := range want {
		if visited[i] != want[i] {
			t.Errorf("at index %d, got %d, want %d", i, visited[i], want[i])
		}
	}
	// Early stop
	visited = nil
	l.ForEach(func(v int) bool {
		visited = append(visited, v)
		return v < 2
	})
	if len(visited) != 2 {
		t.Errorf("expected early stop after 2 elements, got %d", len(visited))
	}
}

func TestForEachPtr(t *testing.T) {
	l := New[int]()
	l.PushBack(10)
	l.PushBack(20)
	var visited []int
	l.ForEachPtr(func(p *int) bool {
		if p == nil {
			t.Errorf("unexpected nil pointer")
			return false
		}
		visited = append(visited, *p)
		return true
	})
	want := []int{10, 20}
	for i := range want {
		if visited[i] != want[i] {
			t.Errorf("at index %d, got %d, want %d", i, visited[i], want[i])
		}
	}
}

func TestRemoveIf(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}
	removed := l.RemoveIf(func(v int) bool { return v%2 == 0 })
	if removed != 2 {
		t.Errorf("expected 2 elements removed, got %d", removed)
	}
	got := l.ToSlice()
	want := []int{1, 3, 5}
	if len(got) != len(want) {
		t.Fatalf("expected %d elements, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at index %d, got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestConcat(t *testing.T) {
	l1 := New[int]()
	l2 := New[int]()
	l1.PushBack(1)
	l1.PushBack(2)
	l2.PushBack(3)
	l2.PushBack(4)
	l1.Concat(l2)
	got := l1.ToSlice()
	want := []int{1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("expected %d elements, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at index %d, got %d, want %d", i, got[i], want[i])
		}
	}

	if !l2.IsEmpty() {
		t.Errorf("expected other list to be empty after Concat")
	}
}

func TestClear(t *testing.T) {
	l := New[int]()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	l.Clear()
	if !l.IsEmpty() {
		t.Errorf("expected list to be empty after Clear")
	}
	if l.Len() != 0 {
		t.Errorf("expected length 0 after Clear, got %d", l.Len())
	}
}

func TestIter(t *testing.T) {
	l := New[int]()
	it := l.Iter()
	if it == nil {
		t.Errorf("expected non-nil iterator for empty list")
	}
	l.PushBack(1)
	it2 := l.Iter()
	if it2 == nil {
		t.Errorf("expected non-nil iterator for non-empty list")
	}
}

func TestSliceAndPtrSlice(t *testing.T) {
	lEven := New[int]()
	for i := 1; i <= 6; i++ {
		lEven.PushBack(i)
	}
	lOdd := New[int]()
	for i := 1; i <= 5; i++ {
		lOdd.PushBack(i)
	}
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
			if len(got) != len(tt.wantVals) {
				t.Fatalf("Slice length = %d, want %d", len(got), len(tt.wantVals))
			}
			for i := range tt.wantVals {
				if got[i] != tt.wantVals[i] {
					t.Errorf("Slice at %d = %d, want %d", i, got[i], tt.wantVals[i])
				}
			}

			gotPtrs := tt.list.PtrSlice(tt.start, tt.end)
			if len(gotPtrs) != len(tt.wantVals) {
				t.Fatalf("PtrSlice length = %d, want %d", len(tt.wantVals), len(gotPtrs))
			}
			for i := range tt.wantVals {
				if *gotPtrs[i] != tt.wantVals[i] {
					t.Errorf("PtrSlice at %d = %d, want %d", i, *gotPtrs[i], tt.wantVals[i])
				}
			}
		})
	}
}

func TestToSliceAndToPtrSlice(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}
	got := l.ToSlice()
	want := []int{1, 2, 3}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ToSlice at %d = %d, want %d", i, got[i], want[i])
		}
	}
	gotPtrs := l.ToPtrSlice()
	for i := range want {
		if *gotPtrs[i] != want[i] {
			t.Errorf("ToPtrSlice at %d = %d, want %d", i, *gotPtrs[i], want[i])
		}
	}
}

func TestToChanAndToPtrChan(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}

	ch := l.ToChan(0)
	var vals []int
	for v := range ch {
		vals = append(vals, v)
	}
	want := []int{1, 2, 3}
	for i := range want {
		if vals[i] != want[i] {
			t.Errorf("ToChan at %d = %d, want %d", i, vals[i], want[i])
		}
	}

	ptrCh := l.ToPtrChan(0)
	var ptrVals []int
	for p := range ptrCh {
		if p == nil {
			t.Errorf("unexpected nil pointer from ToPtrChan")
			continue
		}
		ptrVals = append(ptrVals, *p)
	}
	for i := range want {
		if ptrVals[i] != want[i] {
			t.Errorf("ToPtrChan at %d = %d, want %d", i, ptrVals[i], want[i])
		}
	}
}

func TestIterResetPrevAndResetBackNext(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}
	it := l.Iter()
	it.Reset()
	val, ok := it.Prev()
	if ok {
		t.Errorf("Prev() after Reset should be false, got (%d,true)", val)
	}
	it.ResetBack()
	val, ok = it.Next()
	if ok {
		t.Errorf("Next() after ResetBack should be false, got (%d,true)", val)
	}
}

func TestIterCloneResetAndCurrent(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}
	it := l.Iter()
	if it == nil {
		t.Fatalf("Iter() returned nil")
	}
	// Current should be invalid before Next is called
	val, ok := it.Current()
	if ok {
		t.Errorf("Current() should be false before Next, got (%d,true)", val)
	}
	ptr, ok := it.CurrentPtr()
	if ok || ptr != nil {
		t.Errorf("CurrentPtr() should be (nil,false) before Next, got (%v,%v)", ptr, ok)
	}
	// Advance once
	it.Next()
	val, ok = it.Current()
	if !ok || val != 1 {
		t.Errorf("Current() after Next = (%d,%v), want (1,true)", val, ok)
	}
	ptr, ok = it.CurrentPtr()
	if !ok || ptr == nil || *ptr != 1 {
		t.Errorf("CurrentPtr() after Next = (%v,%v), want (1,true)", ptr, ok)
	}
	// Clone should point to same position
	clone := it.Clone()
	valClone, okClone := clone.Current()
	if !okClone || valClone != 1 {
		t.Errorf("Clone Current() = (%d,%v), want (1,true)", valClone, okClone)
	}
	// Reset should go back to head
	it.Reset()
	it.Next()
	val, ok = it.Current()
	if !ok || val != 1 {
		t.Errorf("Reset then Next Current() = (%d,%v), want (1,true)", val, ok)
	}
	// ResetBack should go to tail
	it.ResetBack()
	val, ok = it.Current()
	if ok {
		t.Errorf("ResetBack then Current() = %d,%v), want (0, false)", val, ok)
	}
	it.Prev() // step back once to get to last item
	it.Next() // step forward once - should not go to tail
	val, ok = it.Current()
	if !ok || val != 3 {
		t.Errorf("ResetBack then Prev Next Current() = (%d,%v), want (3,true)", val, ok)
	}
}

func TestIterEmptyList(t *testing.T) {
	l := New[int]()
	it := l.Iter()
	if it == nil {
		t.Fatalf("Iter() returned nil for empty list")
	}
	val, ok := it.Current()
	if ok || val != 0 {
		t.Errorf("Current() on empty list = (%d,%v), want (0,false)", val, ok)
	}
	ptr, ok := it.CurrentPtr()
	if ok || ptr != nil {
		t.Errorf("CurrentPtr() on empty list = (%v,%v), want (nil,false)", ptr, ok)
	}
	// Reset should still be safe
	it.Reset()
	val, ok = it.Current()
	if ok {
		t.Errorf("Current() after Reset on empty list should be false, got (%d,true)", val)
	}
	// ResetBack should also be safe
	it.ResetBack()
	val, ok = it.Current()
	if ok {
		t.Errorf("Current() after ResetBack on empty list should be false, got (%d,true)", val)
	}
}

func TestIterNavigation(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}
	it := l.Iter()
	if !it.HasNext() {
		t.Errorf("HasNext() at start should be true")
	}
	if it.HasPrev() {
		t.Errorf("HasPrev() at start should be false")
	}
	// Peek should show first element without advancing
	val, ok := it.Peek()
	if !ok || val != 1 {
		t.Errorf("Peek() = (%d,%v), want (1,true)", val, ok)
	}
	// Next should advance into first element
	val, ok = it.Next()
	if !ok || val != 1 {
		t.Errorf("Next() = (%d,%v), want (1,true)", val, ok)
	}
	// Now at element 1 (first non-sentinel), HasPrev should still be false
	if it.HasPrev() {
		t.Errorf("HasPrev() at first element should be false")
	}
	// Advance forward again
	val, ok = it.Next()
	if !ok || val != 2 {
		t.Errorf("Next() = (%d,%v), want (2,true)", val, ok)
	}
	// Now at element 2, HasPrev should be true
	if !it.HasPrev() {
		t.Errorf("HasPrev() should be true at second element")
	}
	// Prev should move back to 1
	val, ok = it.Prev()
	if !ok || val != 1 {
		t.Errorf("Prev() = (%d,%v), want (1,true)", val, ok)
	}
}

func TestIterNavigationEmptyList(t *testing.T) {
	l := New[int]()
	it := l.Iter()

	if it.HasNext() {
		t.Errorf("HasNext() on empty list should be false")
	}
	if it.HasPrev() {
		t.Errorf("HasPrev() on empty list should be false")
	}

	val, ok := it.Next()
	if ok || val != 0 {
		t.Errorf("Next() on empty list = (%d,%v), want (0,false)", val, ok)
	}
	ptr, ok := it.NextPtr()
	if ok || ptr != nil {
		t.Errorf("NextPtr() on empty list = (%v,%v), want (nil,false)", ptr, ok)
	}

	val, ok = it.Prev()
	if ok || val != 0 {
		t.Errorf("Prev() on empty list = (%d,%v), want (0,false)", val, ok)
	}
	ptr, ok = it.PrevPtr()
	if ok || ptr != nil {
		t.Errorf("PrevPtr() on empty list = (%v,%v), want (nil,false)", ptr, ok)
	}

	val, ok = it.Peek()
	if ok || val != 0 {
		t.Errorf("Peek() on empty list = (%d,%v), want (0,false)", val, ok)
	}
	ptr, ok = it.PeekPtr()
	if ok || ptr != nil {
		t.Errorf("PeekPtr() on empty list = (%v,%v), want (nil,false)", ptr, ok)
	}

	val, ok = it.PeekBack()
	if ok || val != 0 {
		t.Errorf("PeekBack() on empty list = (%d,%v), want (0,false)", val, ok)
	}
	ptr, ok = it.PeekBackPtr()
	if ok || ptr != nil {
		t.Errorf("PeekBackPtr() on empty list = (%v,%v), want (nil,false)", ptr, ok)
	}
}

func TestIterRemoveAndInsert(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}
	it := l.Iter()

	// Remove and InsertBefore should fail while at head
	val, ok := it.Remove()
	if ok {
		t.Errorf("Remove() at head sentinel should be false, got (%d,true)", val)
	}
	if it.InsertBefore(99) {
		t.Errorf("InsertBefore() at head sentinel should be false")
	}

	// InsertBefore should succeed while at first item
	it.Next()
	if !it.InsertBefore(77) {
		t.Errorf("InsertBefore() at first element should succeed")
	}
	want := []int{77, 1, 2, 3}
	if got := l.ToSlice(); !slices.Equal(got, want) {
		t.Errorf("InsertBefore failed, got %v, want %v", got, want)
	}

	// Normal path
	if !it.InsertAfter(88) {
		t.Errorf("InsertAfter() at first element should succeed")
	}
	want = []int{77, 1, 88, 2, 3}
	if got := l.ToSlice(); !slices.Equal(got, want) {
		t.Errorf("InsertAfter failed, got %v, want %v", got, want)
	}

	it.Next()
	val, ok = it.Remove()
	if !ok || val != 88 {
		t.Errorf("Remove() = (%d,%v), want (88,true)", val, ok)
	}
	want = []int{77, 1, 2, 3}
	if got := l.ToSlice(); !slices.Equal(got, want) {
		t.Errorf("Remove failed, got %v, want %v", got, want)
	}

	// Remove and InsertAfter should fail while at tail
	it.ResetBack()
	val, ok = it.Remove()
	if ok {
		t.Errorf("Remove() at tail sentinel should be false, got (%d,true)", val)
	}
	if it.InsertAfter(99) {
		t.Errorf("InsertAfter() at tail sentinel should be false")
	}

	// InsertAfter should succeed while at the last element
	it.Prev()
	if !it.InsertAfter(66) {
		t.Errorf("InsertAfter() at last element should succeed")
	}
	want = []int{77, 1, 2, 3, 66}
	if got := l.ToSlice(); !slices.Equal(got, want) {
		t.Errorf("InsertAfter failed, got %v, want %v", got, want)
	}
}

func TestIterForEachAndChannels(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 3; i++ {
		l.PushBack(i)
	}

	// ForEach should visit all items starting from current position
	it := l.Iter()
	var vals []int
	it.ForEach(func(v int) bool {
		vals = append(vals, v)
		return true // continue
	})
	if !slices.Equal(vals, []int{1, 2, 3}) {
		t.Errorf("ForEach got %v, want [1 2 3]", vals)
	}

	// ForEachPtr should visit all items with pointers
	it = l.Iter()
	var ptrVals []int
	it.ForEachPtr(func(p *int) bool {
		ptrVals = append(ptrVals, *p)
		return true
	})
	if !slices.Equal(ptrVals, []int{1, 2, 3}) {
		t.Errorf("ForEachPtr got %v, want [1 2 3]", ptrVals)
	}

	// ForEach can stop early if callback returns false
	it = l.Iter()
	vals = nil
	it.ForEach(func(v int) bool {
		vals = append(vals, v)
		return v < 2 // stop after 2
	})
	if !slices.Equal(vals, []int{1, 2}) {
		t.Errorf("ForEach early stop got %v, want [1 2]", vals)
	}

	// ToChan should produce all remaining items
	it = l.Iter()
	ch := it.ToChan(0) // unbuffered
	var chanVals []int
	for v := range ch {
		chanVals = append(chanVals, v)
	}
	if !slices.Equal(chanVals, []int{1, 2, 3}) {
		t.Errorf("ToChan got %v, want [1 2 3]", chanVals)
	}

	// ToPtrChan should produce pointers
	it = l.Iter()
	ptrCh := it.ToPtrChan(2) // buffered
	var chanPtrVals []int
	for p := range ptrCh {
		chanPtrVals = append(chanPtrVals, *p)
	}
	if !slices.Equal(chanPtrVals, []int{1, 2, 3}) {
		t.Errorf("ToPtrChan got %v, want [1 2 3]", chanPtrVals)
	}
}

func TestIterForEachAndChannelsEmpty(t *testing.T) {
	l := New[int]()
	it := l.Iter()

	// ForEach on empty list should do nothing
	called := false
	it.ForEach(func(v int) bool {
		called = true
		return true
	})
	if called {
		t.Errorf("ForEach on empty list should not call function")
	}

	// ToChan on empty list should yield nothing
	ch := it.ToChan(1)
	if v, ok := <-ch; ok {
		t.Errorf("ToChan on empty list yielded %d", v)
	}

	// ToPtrChan on empty list should yield nothing
	ptrCh := it.ToPtrChan(1)
	if p, ok := <-ptrCh; ok {
		t.Errorf("ToPtrChan on empty list yielded %v", p)
	}
}

func TestIterTakeAndSkip(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}

	// TakeSlice should collect up to n items
	it := l.Iter()
	got := it.TakeSlice(3)
	want := []int{1, 2, 3}
	if !slices.Equal(got, want) {
		t.Errorf("TakeSlice got %v, want %v", got, want)
	}

	// TakePtrSlice should collect pointers
	it = l.Iter()
	ptrs := it.TakePtrSlice(2)
	if len(ptrs) != 2 || *ptrs[0] != 1 || *ptrs[1] != 2 {
		t.Errorf("TakePtrSlice got %v, want [1,2]", ptrs)
	}
}

func TestIterTakeWhileAndSkip(t *testing.T) {
	l := New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}

	// TakeWhile should stop at first non-match
	it := l.Iter()
	got := it.TakeWhile(func(v int) bool { return v < 3 })
	want := []int{1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("TakeWhile got %v, want %v", got, want)
	}

	// TakeWhilePtr should stop at first non-match
	it = l.Iter()
	ptrs := it.TakeWhilePtr(func(p *int) bool { return *p < 4 })
	if len(ptrs) != 3 || *ptrs[2] != 3 {
		t.Errorf("TakeWhilePtr got %v, want [1,2,3]", ptrs)
	}

	// Skip should advance by n
	it = l.Iter()
	it.Skip(2)
	val, ok := it.Next()
	if !ok || val != 3 {
		t.Errorf("Skip(2) then Next = (%d,%v), want (3,true)", val, ok)
	}

	// SkipBack should step back by n
	it.ResetBack()
	it.SkipBack(2)
	val, ok = it.Prev()
	if !ok || val != 3 {
		t.Errorf("SkipBack(2) then Prev = (%d,%v), want (3,true)", val, ok)
	}

	// SkipWhile should advance until predicate fails
	it = l.Iter()
	it.SkipWhile(func(v int) bool { return v < 4 })
	val, ok = it.Next()
	if !ok || val != 4 {
		t.Errorf("SkipWhile then Next = (%d,%v), want (4,true)", val, ok)
	}

	// SkipWhilePtr should advance until predicate fails
	it = l.Iter()
	it.SkipWhilePtr(func(p *int) bool { return *p < 5 })
	val, ok = it.Next()
	if !ok || val != 5 {
		t.Errorf("SkipWhilePtr then Next = (%d,%v), want (5,true)", val, ok)
	}

	// DrainWhile should remove items until predicate fails
	it = l.Iter()
	it.DrainWhile(func(v int) bool { return v < 3 })
	got = l.ToSlice()
	want = []int{3, 4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("DrainWhile modified list to %v, want %v", got, want)
	}

	// DrainWhilePtr should remove items until predicate fails
	l = New[int]()
	for i := 1; i <= 5; i++ {
		l.PushBack(i)
	}
	it = l.Iter()
	it.DrainWhilePtr(func(p *int) bool { return *p < 4 })
	got = l.ToSlice()
	want = []int{4, 5}
	if !slices.Equal(got, want) {
		t.Errorf("DrainWhilePtr modified list to %v, want %v", got, want)
	}
}

func TestIterTakeSkipEmpty(t *testing.T) {
	l := New[int]()
	it := l.Iter()

	if got := it.TakeSlice(3); len(got) != 0 || got == nil {
		t.Errorf("TakeSlice on empty list got %v, want []", got)
	}
	if got := it.TakePtrSlice(3); len(got) != 0 || got == nil {
		t.Errorf("TakePtrSlice on empty list got %v, want []", got)
	}
	if got := it.TakeWhile(func(v int) bool { return true }); len(got) != 0 || got == nil {
		t.Errorf("TakeWhile on empty list got %v, want []", got)
	}
	if got := it.TakeWhilePtr(func(p *int) bool { return true }); len(got) != 0 || got == nil {
		t.Errorf("TakeWhilePtr on empty list got %v, want []", got)
	}

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
	if got, ok := it.Current(); ok {
		t.Errorf("Skip SkipBack SkipWhile DrainWhile DrainWhilePtr on empty, got %v, want %v", got, false)
	}
}

func TestDocExample1(t *testing.T) {
	l := New[int]()
	l.Add(2).Add(4)
	l.PushBack(8)
	l.PushFront(1)
	wantStr := "List[1, 2, 4, 8]"
	if got := l.String(); got != wantStr {
		t.Errorf("String(), got %v, want %v", got, wantStr)
	}

	l = FromSlice[int]([]int{1, 2, 4, 8, 4, 2, 1})
	eq := func(a, b int) bool { return a == b }
	if got := l.IndexOf(2, eq); got != 1 {
		t.Errorf("IndexOf(2), got %v, want %v", got, 1)
	}
	if got := l.LastIndexOf(2, eq); got != 5 {
		t.Errorf("LastIndexof(2), got %v, want %v", got, 5)
	}
	if got := l.IndexOf(65535, eq); got != -1 {
		t.Errorf("IndexOf(65535), got %v, want %v", got, -1)
	}

	l.Iter().Skip(2).ForEachPtr(func(v *int) bool {
		*v *= 2
		return true
	})
	wantStr = "List[1, 2, 8, 16, 8, 4, 2]"
	if s := l.String(); s != wantStr {
		t.Errorf("String(), got %v, want %v", s, wantStr)
	}

	it := l.Iter()
	if v, ok := it.Prev(); v != 0 || ok {
		t.Errorf("Prev() on new iter, got (%v, %v) want (0, false)", v, ok)
	}
	if v, ok := it.Peek(); v != 1 || !ok {
		t.Errorf("Peek() got (%v, %v) want (1, true)", v, ok)
	}
	if v, ok := it.Next(); v != 1 || !ok {
		t.Errorf("Next() got (%v, %v) want (1, true)", v, ok)
	}
	if ok := it.InsertBefore(55); !ok {
		t.Errorf("InsertBefore(55) got %v want true", ok)
	}
	if v, ok := it.Prev(); v != 55 || !ok {
		t.Errorf("Prev() got (%v, %v) want (55, true)", v, ok)
	}
	if v, ok := it.Remove(); v != 55 || !ok {
		t.Errorf("Remove() got (%v, %v) want (55, true)", v, ok)
	}
	if v, ok := it.Next(); v != 1 || !ok {
		t.Errorf("Next() got (%v, %v) want (1, true)", v, ok)
	}
	if v, ok := it.PeekBack(); v != 0 || ok {
		t.Errorf("PeekBack() got (%v, %v) want (0, false)", v, ok)
	}

	wantStr = "List[1, 2, 8, 16, 8, 4, 2]"
	if s := l.String(); s != wantStr {
		t.Errorf("String(), got %v, want %v", s, wantStr)
	}

	v, ok := it.ResetBack().PrevPtr()
	if !ok || *v != 2 {
		t.Errorf("ResetBack() PrevPtr() got (%v, %v) want (2, true)", v, ok)
	}
	*v += 98
	it.InsertAfter(42)
	wantStr = "List[1, 2, 8, 16, 8, 4, 100, 42]"
	if s := l.String(); s != wantStr {
		t.Errorf("String(), got %v, want %v", s, wantStr)
	}

	it.Reset().
		SkipWhile(func(v int) bool {
			return v < 16
		}).
		SkipBack(1).
		DrainWhile(func(v int) bool {
			return v != 4
		})

	wantSlice := []int{1, 2, 4, 100, 42}
	if s := l.ToSlice(); !slices.Equal(s, wantSlice) {
		t.Errorf("String(), got %v, want %v", s, wantSlice)
	}
	wantSlice = []int{2, 4, 100}
	if s := it.Reset().Skip(1).TakeSlice(3); !slices.Equal(s, wantSlice) {
		t.Errorf("String(), got %v, want %v", s, wantSlice)
	}
}
