package unsafeutil

import (
	"testing"
	"unsafe"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestAtIntSlice(t *testing.T) {
	s := []int{10, 20, 30, 40}
	base := unsafe.SliceData(s)

	for i := range s {
		got := *At(base, i)
		want := s[i]
		AssertEq(t, want, got)
	}
}

func TestAtStructSlice(t *testing.T) {
	type Point struct{ X, Y, Z int64 }
	s := []Point{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	base := unsafe.SliceData(s)

	for i := range s {
		got := *At(base, i)
		want := s[i]
		AssertEq(t, want, got)
	}
}
