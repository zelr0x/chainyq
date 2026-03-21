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

package seq_test

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"testing"

	. "github.com/zelr0x/chainyq/internal/testutil"
	"github.com/zelr0x/chainyq/seq"
)

func TestSeqNewFromSliceGen(t *testing.T) {
	xs := SliceFromRangeExcl(t, 0, 20)
	want := slices.Clone(xs)
	i := 0
	s := seq.New(func() (int, bool) {
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
	s := seq.New(func() (int, bool) {
		i++
		return i, true
	})
	got := s.Take(5).ToSlice()
	want := SliceFromRangeIncl(t, 1, 5)
	AssertSliceEq(t, want, got)
}

func TestSeqNewTakeFromInfRandomGen(t *testing.T) {
	rngSeq := seq.New(func() (int, bool) {
		return rand.Intn(100), true
	})
	AssertEq(t, 5, rngSeq.Take(5).Count(), "take rng val count eq")
}

func TestFromAndToSlice(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := seq.FromSlice(want).ToSlice()
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
	readCh := seq.FromChan(ch).ToChan(0)
	got := make([]int, 0, 20)
	for v := range readCh {
		got = append(got, v)
	}
	AssertSliceEq(t, want, got, "FromChan ToChan eq")
}

func TestFilter(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	s := seq.FromSlice(want)
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
			s := seq.FromSlice(SliceFromRangeIncl(t, 1, 10))
			got := s.Take(tt.n).ToSlice()
			AssertSliceEq(t, tt.want, got)
		})
	}
}

func TestSkip(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	k := len(want) / 2
	s := seq.FromSlice(want)
	got := s.Skip(k).ToSlice()
	AssertSliceEq(t, want[k:], got, fmt.Sprintf("Skip(%d)", k))
}

func TestSkipWhile(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 10)
	k := len(want) / 2
	s := seq.FromSlice(want)
	got := s.SkipWhile(func(v int) bool {
		return v < k
	}).ToSlice()
	AssertSliceEq(t, want[k:], got, fmt.Sprintf("SkipWhile(v < %v)", k))
}

func TestMap(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 3)
	s := seq.FromSlice(want)
	got := seq.Map(s, func(x int) string {
		return strconv.Itoa(x)
	}).ToSlice()
	AssertSliceEq(t, []string{"0", "1", "2"}, got)
}

func TestFlatMap(t *testing.T) {
	s := seq.FromSlice([]int{1, 2, 3})
	flat := seq.FlatMap(s, func(x int) seq.Seq[int] {
		return seq.FromSlice([]int{x, x * 10})
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
			s := seq.FromSlice(tt.input)
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
			s := seq.FromSlice(tt.input)
			AssertEq(t, tt.want, s.CountU64())
		})
	}
}

func TestForEach(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := make([]int, 0, len(want))
	seq.FromSlice(want).
		ForEach(func(x int) {
			got = append(got, x)
		})
	AssertSliceEq(t, want, got)
}

func TestForEachUntil(t *testing.T) {
	want := SliceFromRangeExcl(t, 0, 20)
	got := make([]int, 0, len(want))
	k := len(want) / 2
	seq.FromSlice(want).
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
	seq.FromSlice(want).
		ForEachIndexed(func(i int, x int) bool {
			got = append(got, x)
			return i < k
		})
	AssertSliceEq(t, want[:k+1], got)
}

func TestToMap(t *testing.T) {
	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
	m := seq.ToMap(s,
		func(s string) string { return string(s[0]) },
		func(s string) string { return s },
	)
	got, ok := m["a"]
	AssertEqOk(t, "apricot", got, ok, "expected apricot to overwrite apple")
	got, ok = m["b"]
	AssertEqOk(t, "banana", got, ok)
}

func TestToMapMerge(t *testing.T) {
	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
	m := seq.ToMapMerge(s,
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
	s := seq.FromSlice([]string{"apple", "apricot", "banana"})
	groups := seq.GroupByHint(s,
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
	s := seq.FromSlice([]int{1, 2, 3})
	sum := seq.Fold(s, 0, func(acc, x int) int {
		return acc + x
	})
	AssertEq(t, 6, sum)
}

func TestFoldr(t *testing.T) {
	s := seq.FromSlice([]int{1, 2, 3})
	res := seq.Foldr(s, 0, func(x int, rest func() int) int {
		return x + rest()
	})
	AssertEq(t, 6, res)

	// Foldr should evaluate as 1 - (2 - (3 - 0)) = 2
	s = seq.FromSlice([]int{1, 2, 3})
	res = seq.Foldr(s, 0, func(x int, rest func() int) int {
		return x - rest()
	})
	AssertEq(t, 2, res)
}
