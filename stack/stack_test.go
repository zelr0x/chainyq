package stack

import (
	"fmt"
	"testing"
	"unsafe"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestString(t *testing.T) {
	var d *Stack[int]
	AssertEq(t, "nil", d.String())
	d = New[int]()
	AssertEq(t, "Stack[]", d.String())
	slice := SliceFromRangeIncl(t, 1, 10)
	for _, v := range slice {
		d.PushBack(v)
	}
	want := ReversedSlice(slice)
	AssertEq(t, fmt.Sprintf("Stack%v", want), d.String())
}

func TestGoString(t *testing.T) {
	var d *Stack[int]
	AssertEq(t, "nil", d.String())
	d = New[int]()
	AssertEq(t, "Stack[int]{}", fmt.Sprintf("%#v", d))
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)
	AssertEq(t, "Stack[int]{3, 2, 1}", fmt.Sprintf("%#v", d))
}

func TestEquals(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name string
		a    *Stack[int]
		b    *Stack[int]
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

func TestIntStackOps(t *testing.T) {
	cases := []struct {
		name      string
		input     []int
		expectOk  bool
		expectVal int
	}{
		{"empty", []int{}, false, 0},
		{"single", []int{42}, true, 42},
		{"odd", SliceFromRangeIncl(t, 1, 5), true, 5},
		{"even", SliceFromRangeIncl(t, 1, 6), true, 6},
		{"large", SliceFromRangeExcl(t, 0, 10000), true, 9999},
	}
	for _, tt := range cases {
		t.Run(tt.name+"_Pop", func(t *testing.T) {
			s := New[int]()
			for _, x := range tt.input {
				s.Push(x)
			}
			v, ok := s.Pop()
			AssertEq(t, tt.expectOk, ok, "Pop ok mismatch")
			AssertEq(t, tt.expectVal, v, "Pop value mismatch")
		})
		t.Run(tt.name+"_Peek", func(t *testing.T) {
			s := New[int]()
			for _, x := range tt.input {
				s.Push(x)
			}
			v, ok := s.Peek()
			AssertEq(t, tt.expectOk, ok, "Peek ok mismatch")
			AssertEq(t, tt.expectVal, v, "Peek value mismatch")
		})
		t.Run(tt.name+"_PeekPtr", func(t *testing.T) {
			s := New[int]()
			for _, x := range tt.input {
				s.Push(x)
			}
			ptr, ok := s.PeekPtr()
			AssertEq(t, tt.expectOk, ok, "PeekPtr ok mismatch")
			if tt.expectOk {
				AssertEq(t, tt.expectVal, *ptr, "PeekPtr value mismatch")
			} else {
				AssertEq(t, (*int)(nil), ptr, "PeekPtr nil mismatch")
			}
		})
	}
}

func TestStringStackOps(t *testing.T) {
	cases := []struct {
		name      string
		input     []string
		expectOk  bool
		expectVal string
	}{
		{"empty", []string{}, false, ""},
		{"single", []string{"x"}, true, "x"},
		{"odd", []string{"a", "b", "c"}, true, "c"},
		{"even", []string{"a", "b", "c", "d"}, true, "d"},
	}
	for _, tt := range cases {
		t.Run(tt.name+"_Pop", func(t *testing.T) {
			s := New[string]()
			for _, x := range tt.input {
				s.Push(x)
			}
			v, ok := s.Pop()
			AssertEq(t, tt.expectOk, ok, "Pop ok mismatch")
			AssertEq(t, tt.expectVal, v, "Pop value mismatch")
		})
		t.Run(tt.name+"_Peek", func(t *testing.T) {
			s := New[string]()
			for _, x := range tt.input {
				s.Push(x)
			}
			v, ok := s.Peek()
			AssertEq(t, tt.expectOk, ok, "Peek ok mismatch")
			AssertEq(t, tt.expectVal, v, "Peek value mismatch")
		})
		t.Run(tt.name+"_PeekPtr", func(t *testing.T) {
			s := New[string]()
			for _, x := range tt.input {
				s.Push(x)
			}
			ptr, ok := s.PeekPtr()
			AssertEq(t, tt.expectOk, ok, "PeekPtr ok mismatch")
			if tt.expectOk {
				AssertEq(t, tt.expectVal, *ptr, "PeekPtr value mismatch")
			} else {
				AssertEq(t, (*string)(nil), ptr, "PeekPtr nil mismatch")
			}
		})
	}
}

func TestStackExample1(t *testing.T) {
	s := New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if v, ok := s.PeekPtr(); ok {
		*v = 50
	}
	AssertSliceEq(t, []int{50, 2, 1}, s.ToSlice())
	AssertSliceEq(t, []int{1, 2, 50}, s.UnwrapCopy())
	backing := s.UnwrapUnsafe()
	AssertSliceEq(t, []int{1, 2, 50}, backing)
	want := (*[2]uintptr)(unsafe.Pointer(&s.b))[0]
	got := (*[2]uintptr)(unsafe.Pointer(&backing))[0]
	AssertEq(t, want, got)
}
