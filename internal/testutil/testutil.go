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

package testutil

import (
	"slices"
	"testing"
)

func AssertEq[T comparable](
	t *testing.T,
	want T,
	got T,
	msg ...string,
) {
	t.Helper()
	if got == want {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Errorf("want %v, got %v", want, got)
	}
}

func AssertEqOk[T comparable](
	t *testing.T,
	want T,
	got T,
	gotOK bool,
	msg ...string,
) {
	t.Helper()
	if got == want {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want (%v, true), got (%v, %v)", msg[0], want, got, gotOK)
	} else {
		t.Errorf("want (%v, true), got (%v, %v)", want, got, gotOK)
	}
}

func AssertNotNil[T any](
	t *testing.T,
	got *T,
	msg ...string,
) {
	t.Helper()
	if got != nil {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: expected not nil", msg[0])
	} else {
		t.Error("expected not nil")
	}
}

func AssertTrue(
	t *testing.T,
	got bool,
	msg ...string,
) {
	t.Helper()
	if got {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want true, got %v", msg[0], got)
	} else {
		t.Error("want true, got false")
	}
}

func AssertFalse(
	t *testing.T,
	got bool,
	msg ...string,
) {
	t.Helper()
	if !got {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want false, got true", msg[0])
	} else {
		t.Error("want false, got true")
	}
}

func AssertZero[T comparable](
	t *testing.T,
	got T,
	msg ...string,
) {
	t.Helper()
	var want T
	if got == want {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Errorf("want %v, got %v", want, got)
	}
}

func AssertZeroFalse[T comparable](
	t *testing.T,
	got T,
	gotOK bool,
	msg ...string,
) {
	t.Helper()
	var want T
	if !gotOK && got == want {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want (%v, false), got (%v, %v)", msg[0], want, got, gotOK)
	} else {
		t.Errorf("want (%v, false), got (%v, %v)", want, got, gotOK)
	}
}

func AssertPtrSliceEq[T comparable](
	t *testing.T,
	want []T,
	got []*T,
	msg ...string,
) {
	t.Helper()
	if len(got) != len(want) {
		if len(msg) > 0 {
            t.Errorf("%s (len): want %v, got %v", msg[0], len(want), len(got))
        } else {
            t.Errorf("(len): want %v, got %v", len(want), len(got))
        }
	}
	deref := DerefSlice(t, got)
	if !slices.Equal(deref, want) {
		if len(msg) > 0 {
			t.Errorf("%s: want %v, got %v", msg[0], want, deref)
        } else {
            t.Errorf("want %v, got %v", want, deref)
        }
	}
}

func AssertSliceEq[T comparable](
	t *testing.T,
	want []T,
	got []T,
	msg ...string,
) {
	t.Helper()
	if slices.Equal(got, want) {
		return
	}
	if len(msg) > 0 {
		t.Errorf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Errorf("want %v, got %v", want, got)
	}
}

func ReversedSlice[T any](slice []T) []T {
	revSlice := make([]T, len(slice))
	copy(revSlice, slice)
	slices.Reverse(revSlice)
	return revSlice
}

func DerefSlice[T any](t *testing.T, slice []*T) []T {
	t.Helper()
	if len(slice) == 0 {
		return []T{}
	}
	res := make([]T, 0, len(slice))
	for _, v := range slice {
		if v == nil {
			var zero T
			res = append(res, zero)
			continue
		}
		res = append(res, *v)
	}
	return res
}

func SliceFromRangeExcl(t *testing.T, from, to int) []int {
	t.Helper()
	ValidRangeOrDie(t, from, to)
	res := make([]int, 0, to-from)
	for i := from; i < to; i++ {
		res = append(res, i)
	}
	return res
}

func SliceFromRangeIncl(t *testing.T, from, to int) []int {
	t.Helper()
	ValidRangeOrDie(t, from, to)
	res := make([]int, 0, to-from+1)
	for i := from; i <= to; i++ {
		res = append(res, i)
	}
	return res
}

// ValidRangeOrDie errors on from < 0 OR to < 1 OR from >= to
func ValidRangeOrDie(t *testing.T, from, to int) {
	t.Helper()
	if from < 0 {
		t.Error("from < 0")
	}
	if to < 1 {
		t.Error("to < 1")
	}
	if from >= to {
		t.Error("from >= to")
	}
}
