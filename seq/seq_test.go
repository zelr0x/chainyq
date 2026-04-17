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

package seq

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"testing"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestSeqNewFromSliceGen(t *testing.T) {
	xs := SliceFromRangeExcl(t, 0, 20)
	want := slices.Clone(xs)
	i := 0
	s := New(func() (int, bool) {
		if i >= len(xs) {
			return 0, false
		}
		v := xs[i]
		i++
		return v, true
	})
	got := s.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestSeqNewTakeFromInfCounterGen(t *testing.T) {
	i := 0
	s := New(func() (int, bool) {
		i++
		return i, true
	})
	got := s.Take(5).ToSlice()
	want := SliceFromRangeIncl(t, 1, 5)
	AssertSliceEq(t, want, got)
}

func TestSeqNewTakeFromInfRandomGen(t *testing.T) {
	rngSeq := New(func() (int, bool) {
		return rand.Intn(100), true
	})
	AssertEq(t, 5, rngSeq.Take(5).Count(), "take rng val count eq")
}

func TestFromAndToSlice(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := FromSlice(want).ToSlice()
	AssertSliceEq(t, want, got, "FromSlice ToSlice eq")
}

func TestFromChanToChan(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range want {
			ch <- v
		}
	}()
	readCh := FromChan(ch).ToChan(0)
	got := make([]int, 0, 20)
	for v := range readCh {
		got = append(got, v)
	}
	AssertSliceEq(t, want, got, "FromChan ToChan eq")
}

func TestNext(t *testing.T) {
	xs := SliceFromRangeExcl(t, 0, 20)
	s := FromSlice(xs)
	for range xs {
		v, ok := s.Next()
		AssertEqOk(t, v, v, ok)
	}
}

func TestFilter(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	s := FromSlice(want)
	got := s.Filter(func(x int) bool {
		return x%2 == 0
	}).ToSlice()
	AssertSliceEq(t, []int{0, 2, 4, 6, 8}, got)
}

func TestTake(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []int
	}{
		{"negative take", -5, []int{}},
		{"zero take", 0, []int{}},
		{"take one", 1, []int{1}},
		{"take more than length", 100, SliceFromRangeIncl(t, 1, 10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FromSlice(SliceFromRangeIncl(t, 1, 10))
			got := s.Take(tt.n).ToSlice()
			AssertSliceEq(t, tt.want, got)
		})
	}
}

func TestSkipNeg(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	s := FromSlice(want)
	got := s.Skip(-1).ToSlice()
	AssertSliceEq(t, want, got)
}

func TestSkipFromSlice(t *testing.T) {
	want := SliceFromRangeIncl(t, 1, 40)
	k := len(want) / 2
	s := FromSlice(want).Skip(k).Take(k)
	got := s.ToSlice()
	AssertSliceEq(t, want[k:], got, fmt.Sprintf("Skip(%d)", k))
	AssertNotNil(t, s.remaining, "seq.remaining should be not nil")
	AssertEq(t, 0, *s.remaining, "seq.remaining should be zeroed")
}

func TestSkipBiggerThanLenFromSlice(t *testing.T) {
	want := SliceFromRangeIncl(t, 1, 40)
	s := FromSlice(want).Skip(41)
	got := s.ToSlice()
	AssertSliceEq(t, []int{}, got, "Skip(41) should terminate early")
	AssertRemainingNilOrZero(t, s)
}

func TestSkipInIncrementsFromSlice(t *testing.T) {
	want := SliceFromRangeIncl(t, 1, 40)
	s := FromSlice(want).Skip(20).Skip(21)
	got := s.ToSlice()
	AssertSliceEq(t, []int{}, got, "Skip(20) + Skip(21) should terminate early")
	AssertRemainingNilOrZero(t, s)
}

func TestSkipInf(t *testing.T) {
	want := SliceFromRangeIncl(t, 1, 40)
	i := 0
	s := New(func() (int, bool) {
		i++
		return i, true
	})
	AssertNil(t, s.remaining)
	AssertZero(t, s.maxLen)
	k := len(want) / 2
	s = s.Skip(k).Take(k)
	got := s.ToSlice()
	AssertFalse(t, s.IsExactSized(), "must be Sized but not ExactSized after Take on inf seq (1)")
	AssertEq(t, uint64(k), s.maxLen, "must be Sized but not ExactSized after Take on inf seq (2)")
	AssertSliceEq(t, want[k:], got, fmt.Sprintf("Skip(%d)", k))
}

func TestSkipWhile(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	k := len(want) / 2
	s := FromSlice(want)
	got := s.SkipWhile(func(v int) bool {
		return v < k
	}).ToSlice()
	AssertSliceEq(t, want[k:], got, fmt.Sprintf("SkipWhile(v < %v)", k))
}

func TestSkipWhileMatchesWithGap(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	s := FromSlice(want)
	got := s.SkipWhile(func(v int) bool {
		return v < 10 || v > 15
	}).ToSlice()
	AssertSliceEq(t, want[10:], got, "SkipWhile(v < 10 && v > 15)")
}

func TestMap(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 3)
	s := FromSlice(want)
	got := Map(s, func(x int) string {
		return strconv.Itoa(x)
	}).ToSlice()
	AssertSliceEq(t, []string{"0", "1", "2"}, got)
}

func TestFlatMap(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	flat := FlatMap(s, func(x int) Seq[int] {
		return FromSlice([]int{x, x * 10})
	}).ToSlice()
	AssertSliceEq(t, []int{1, 10, 2, 20, 3, 30}, flat)
}

func TestSeqCount(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  int
	}{
		{"empty slice", []int{}, 0},
		{"non-empty slice", make([]int, 150), 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FromSlice(tt.input)
			AssertEq(t, tt.want, s.Count())
		})
	}
}

func TestSeqCountU64(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  uint64
	}{
		{"empty slice", []int{}, uint64(0)},
		{"non-empty slice", make([]int, 150), uint64(150)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FromSlice(tt.input)
			AssertEq(t, tt.want, s.CountU64())
		})
	}
}

func TestForEach(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := make([]int, 0, len(want))
	FromSlice(want).
		ForEach(func(x int) {
			got = append(got, x)
		})
	AssertSliceEq(t, want, got)
}

func TestForEachUntil(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := make([]int, 0, len(want))
	k := len(want) / 2
	FromSlice(want).
		ForEachUntil(func(x int) bool {
			got = append(got, x)
			return x < k
		})
	AssertSliceEq(t, want[:k+1], got)
}

func TestForEachIndexed(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := make([]int, 0, len(want))
	k := len(want) / 2
	FromSlice(want).
		ForEachIndexed(func(i int, x int) bool {
			got = append(got, x)
			return i < k
		})
	AssertSliceEq(t, want[:k+1], got)
}

func TestIterRange(t *testing.T) {
	even := SliceFromRangeIncl(t, 1, 6)
	odd := even[0 : len(even)-1]
	tests := []struct {
		name  string
		slice []int
		start int
		end   int
	}{
		{"empty slice", []int{}, 0, 3},
		{"full slice even length", even, 0, 6},
		{"full slice odd length", odd, 0, 5},
		{"partial slice middle", even, 2, 5},
		{"slice with start=0 end=0", even, 0, 0},
		{"slice with start=end", even, 3, 3},
		{"slice with start>end", even, 4, 2},
		{"slice with negative start", even, -2, 3},
		{"slice with negative end", even, 0, -1},
		{"slice with end beyond length", odd, 0, 10},
		{"slice with start beyond length", odd, 10, 15},
		{"slice covering tail", even, 4, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := tt.slice
			start := tt.start
			end := tt.end
			var want []int
			if start >= 0 && start < len(slice) && end > start {
				want = slice[start:min(end, len(slice))]
			} else {
				want = []int{}
			}

			got := make([]int, 0, len(want))
			for v := range FromSlice(slice).IterRange(start, end) {
				got = append(got, v)
			}
			AssertSliceEq(t, want, got)

			if len(want) > 0 {
				k := len(want) / 2
				want := want[:k]
				got := make([]int, 0, k)
				i := 0
				for v := range FromSlice(slice).IterRange(start, end) {
					got = append(got, v)
					i++
					if i == k {
						break
					}
				}
				AssertSliceEq(t, want, got)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	s := FromSlice([]string{"apple", "apricot", "banana"})
	m := ToMap(s,
		func(s string) string { return string(s[0]) },
		func(s string) string { return s },
	)
	got, ok := m["a"]
	AssertEqOk(t, "apricot", got, ok, "expected apricot to overwrite apple")
	got, ok = m["b"]
	AssertEqOk(t, "banana", got, ok)
}

func TestToMapMerge(t *testing.T) {
	s := FromSlice([]string{"apple", "apricot", "banana"})
	m := ToMapMerge(s,
		func(s string) string { return string(s[0]) },
		func(s string) string { return s },
		func(a, b string) string { return a },
		2,
	)
	got, ok := m["a"]
	AssertEqOk(t, "apple", got, ok, "expected no overwrite")
	got, ok = m["b"]
	AssertEqOk(t, "banana", got, ok)
}

func TestGroupByHint(t *testing.T) {
	s := FromSlice([]string{"apple", "apricot", "banana"})
	groups := GroupByHint(s,
		func(s string) string { return string(s[0]) },
		2,
		2,
	)
	a := groups["a"]
	b := groups["b"]
	AssertSliceEq(t, []string{"apple", "apricot"}, a, "GroupByHint aWords")
	AssertSliceEq(t, []string{"banana"}, b, "GroupByHint bWords")
}

func TestFold(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	sum := Fold(s, 0, func(acc, x int) int {
		return acc + x
	})
	AssertEq(t, 6, sum)
}

func TestFoldr(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	res := Foldr(s, 0, func(x int, rest func() int) int {
		return x + rest()
	})
	AssertEq(t, 6, res)

	// Foldr should evaluate as 1 - (2 - (3 - 0)) = 2
	s = FromSlice([]int{1, 2, 3})
	res = Foldr(s, 0, func(x int, rest func() int) int {
		return x - rest()
	})
	AssertEq(t, 2, res)
}

// ----- Helpers -----

// AssertRemainingNilOrZero is useful because we don't assume anything
// about early-termination impl, so both nil and 0 are acceptable.
func AssertRemainingNilOrZero(t *testing.T, s Seq[int]) {
	AssertTrue(t, s.remaining == nil || *s.remaining == 0,
		"seq.remaining should either be nil or 0")
}
