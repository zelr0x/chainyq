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
	"math/rand"
	"reflect"
	"slices"
	"testing"
	"unsafe"
)

const seed int64 = 31337

func AssertSameSlice[T any](
	t *testing.T,
	want []T,
	got []T,
	msg ...string,
) {
	t.Helper()
	if want == nil {
		AssertNil(t, got, msg...)
		return
	}
	AssertNotNil(t, got, msg...)
	wantSD := unsafe.SliceData(want) // #nosec G103
	gotSD := unsafe.SliceData(got)   // #nosec G103
	AssertEq(t, wantSD, gotSD, msg...)
}

func AssertNotSameSlice[T any](
	t *testing.T,
	notWant []T,
	got []T,
	msg ...string,
) {
	t.Helper()
	if notWant == nil {
		AssertNotNil(t, got, msg...)
		return
	}
	if got == nil {
		return
	}
	AssertNotNil(t, got, msg...)
	wantSD := unsafe.SliceData(notWant) // #nosec G103
	gotSD := unsafe.SliceData(got)      // #nosec G103
	AssertNotEq(t, wantSD, gotSD, msg...)
}

func AssertEqual[T any](
	t *testing.T,
	want T,
	got T,
	msg ...string,
) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	if len(msg) > 0 {
		t.Fatalf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Fatalf("want %v, got %v", want, got)
	}
}

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
		t.Fatalf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func AssertNotEq[T comparable](
	t *testing.T,
	notWant T,
	got T,
	msg ...string,
) {
	t.Helper()
	if got != notWant {
		return
	}
	if len(msg) > 0 {
		t.Fatalf("%s: want not equal to %v, got %v", msg[0], notWant, got)
	} else {
		t.Fatalf("want not equal to %v, got %v", notWant, got)
	}
}

func AssertCommaOk[T comparable](
	t *testing.T,
	want T,
	wantOK bool,
	got T,
	gotOK bool,
	msg ...string,
) {
	t.Helper()
	if wantOK {
		AssertEqOk(t, want, got, gotOK, msg...)
	} else {
		AssertZeroFalse(t, got, gotOK, msg...)
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
		t.Fatalf("%s: want (%v, true), got (%v, %v)", msg[0], want, got, gotOK)
	} else {
		t.Fatalf("want (%v, true), got (%v, %v)", want, got, gotOK)
	}
}

func AssertNil[T any](
	t *testing.T,
	got T,
	msg ...string,
) {
	t.Helper()
	if IsNil(got) {
		return
	}
	if len(msg) > 0 {
		t.Fatalf("%s: expected nil", msg[0])
	} else {
		t.Fatal("expected nil")
	}
}

func AssertNotNil[T any](
	t *testing.T,
	got T,
	msg ...string,
) {
	t.Helper()
	if !IsNil(got) {
		return
	}
	if len(msg) > 0 {
		t.Fatalf("%s: expected not nil", msg[0])
	} else {
		t.Fatal("expected not nil")
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
		t.Fatalf("%s: want true, got %v", msg[0], got)
	} else {
		t.Fatal("want true, got false")
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
		t.Fatalf("%s: want false, got true", msg[0])
	} else {
		t.Fatal("want false, got true")
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
		t.Fatalf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Fatalf("want %v, got %v", want, got)
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
		t.Fatalf("%s: want (%v, false), got (%v, %v)", msg[0], want, got, gotOK)
	} else {
		t.Fatalf("want (%v, false), got (%v, %v)", want, got, gotOK)
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
			t.Fatalf("%s (len): want %v, got %v", msg[0], len(want), len(got))
		} else {
			t.Fatalf("(len): want %v, got %v", len(want), len(got))
		}
	}
	deref := DerefSlice(t, got)
	if !slices.Equal(deref, want) {
		if len(msg) > 0 {
			t.Fatalf("%s: want %v, got %v", msg[0], want, deref)
		} else {
			t.Fatalf("want %v, got %v", want, deref)
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
		t.Fatalf("%s: want %v, got %v", msg[0], want, got)
	} else {
		t.Fatalf("want %v, got %v", want, got)
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

// SliceFromRangeExclWithGaps is the same as [SliceFromRangeExcl] but
// excludes the ones in exclude. Linear search - will be slow for many excludes.
func SliceFromRangeExclWithGaps(t *testing.T, from, to int, exclude ...int) []int {
	t.Helper()
	ValidRangeOrDie(t, from, to)
	res := make([]int, 0, to-from)
	for i := from; i < to; i++ {
		if slices.Contains(exclude, i) {
			continue
		}
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

func RandomIntSlice(t *testing.T, intn int) []int {
	t.Helper()
	return RandomIntSliceN(t, intn, intn)
}

func RandomIntSliceN(t *testing.T, size, intn int) []int {
	t.Helper()
	rand := rand.New(rand.NewSource(seed))
	a := make([]int, size)
	for i := range size {
		a[i] = rand.Intn(intn)
	}
	return a
}

func IsNil[T any](x T) bool {
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	}
	return false
}

func IsSameSlice[T any](
	a []T,
	b []T,
) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	aSD := unsafe.SliceData(a) // #nosec G103
	bSD := unsafe.SliceData(b) // #nosec G103
	return aSD == bSD
}
