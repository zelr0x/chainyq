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

// Package deque provides [Deque] - a segmented-array-based deque.
package deque

import (
	"fmt"
	"math/bits"
	"math/rand"
	"testing"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

var defCfg DequeCfg = DequeCfg{
	BlockSize: 4,
	FrontCap:  2,
	BackCap:   2,
}

var defCfgPooled = func() DequeCfg {
	copy := defCfg
	copy.Pooled = true
	// copy.PoolPrealloc = 16
	return copy
}()

const defCfgInitCap int = 4 * 4

// This is for debugging/inspecting pools.
// func TestChurnPooled(t *testing.T) {
// 	n := 100_000_000
// 	k := 1000
// 	a := RandomIntSliceN(t, k, 4)
// 	maxLen := max(1024, n/2)
// 	d := NewPooled[int]()
// 	for i := range n {
// 		switch a[i%k] {
// 		case 0:
// 			d.PushBack(i)
// 		case 1:
// 			d.PushFront(i)
// 		case 2:
// 			d.PopBack()
// 		case 3:
// 			d.PopFront()
// 		}
// 		if d.Len() > maxLen {
// 			d.Clear()
// 		}
// 	}
// 	AssertTrue(t, d.Len() > -1, "avoid optimizing away")
// }

func TestSuggestBlockSize(t *testing.T) {
	type tiny struct{}
	type small struct{ _ int32 }
	type medium struct{ _, _ int64 }
	type big struct{ _ [33]byte } // covers misaligned branch: if rem != 0 ...
	tests := []struct {
		name string
		fn   func() int
		want func(int) bool
	}{
		{
			"ZST",
			func() int { return SuggestBlockSize[tiny]() },
			func(v int) bool { return v == minBlockSize },
		},
		{
			"small type int32",
			func() int { return SuggestBlockSize[small]() },
			func(v int) bool {
				return v >= minBlockSize && v&(v-1) == 0
			},
		},
		{
			"medium type int64 pair",
			func() int { return SuggestBlockSize[medium]() },
			func(v int) bool {
				return v >= minBlockSize && v&(v-1) == 0
			},
		},
		{
			"big struct bytes",
			func() int { return SuggestBlockSize[big]() },
			func(v int) bool {
				return v >= minBlockSize && v&(v-1) == 0
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			want := tt.want(got)
			AssertTrue(t, want, fmt.Sprintf("SuggestBlockSize() = %d", got))
		})
	}
}

func TestNewValue(t *testing.T) {
	suggestedIntBlockSize := SuggestBlockSize[int]()
	tests := []struct {
		name        string
		cfg         DequeCfg
		wantBlkSize int
		wantShift   uint
		wantFront   int
		wantTotal   int
	}{
		{
			name:        "default config",
			cfg:         DequeCfg{},
			wantBlkSize: suggestedIntBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(suggestedIntBlockSize))), // #nosec G115
			wantFront:   defSideCapBlocks,
			wantTotal:   2 * defSideCapBlocks,
		},
		{
			name:        "block size below min",
			cfg:         DequeCfg{BlockSize: minBlockSize - 1},
			wantBlkSize: minBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(minBlockSize))), // #nosec G115
			wantFront:   defSideCapBlocks,
			wantTotal:   2 * defSideCapBlocks,
		},
		{
			name:        "block size above max",
			cfg:         DequeCfg{BlockSize: maxBlockSize + 1},
			wantBlkSize: maxBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(maxBlockSize))), // #nosec G115
			wantFront:   defSideCapBlocks,
			wantTotal:   2 * defSideCapBlocks,
		},
		{
			name:        "block size rounded to next pow2",
			cfg:         DequeCfg{BlockSize: (maxBlockSize << 1) + 1},
			wantBlkSize: maxBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(maxBlockSize))), // #nosec G115
			wantFront:   defSideCapBlocks,
			wantTotal:   2 * defSideCapBlocks,
		},
		{
			name:        "front capacity specified",
			cfg:         DequeCfg{FrontCap: 10000},
			wantBlkSize: suggestedIntBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(suggestedIntBlockSize))), // #nosec G115
			wantFront:   blocksForCapCeil(suggestedIntBlockSize, 10000),
			wantTotal:   defSideCapBlocks + blocksForCapCeil(suggestedIntBlockSize, 10000),
		},
		{
			name:        "back capacity specified",
			cfg:         DequeCfg{BackCap: 10000},
			wantBlkSize: suggestedIntBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(suggestedIntBlockSize))), // #nosec G115
			wantFront:   defSideCapBlocks,
			wantTotal:   defSideCapBlocks + blocksForCapCeil(suggestedIntBlockSize, 10000),
		},
		{
			name:        "front and back capacity specified",
			cfg:         DequeCfg{FrontCap: 2000, BackCap: 8000},
			wantBlkSize: suggestedIntBlockSize,
			wantShift:   uint(bits.TrailingZeros(uint(suggestedIntBlockSize))), // #nosec G115
			wantFront:   blocksForCapCeil(suggestedIntBlockSize, 2000),
			wantTotal: blocksForCapCeil(suggestedIntBlockSize, 2000) +
				blocksForCapCeil(suggestedIntBlockSize, 8000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewValue[int](tt.cfg)
			AssertEq(t, tt.wantBlkSize, d.blockSize)
			AssertEq(t, tt.wantShift, d.blockShift)
			AssertEq(t, tt.wantFront, d.initCfg.frontBlock)
			AssertEq(t, tt.wantFront, d.initCfg.backBlock)
			AssertEq(t, tt.wantTotal, d.initCfg.totalBlocks)
		})
	}
}

func TestString(t *testing.T) {
	var d *Deque[int]
	AssertEq(t, "nil", d.String())
	d = WithCfg[int](defCfg)
	AssertEq(t, "Deque[]", d.String())
	slice := SliceFromRangeIncl(t, 1, 1+defCfgInitCap)
	for _, v := range slice {
		d.PushBack(v)
	}
	AssertEq(t, fmt.Sprintf("Deque%v", slice), d.String())
}

func TestGoString(t *testing.T) {
	var d *Deque[int]
	AssertEq(t, "nil", d.String())
	d = New[int]()
	AssertEq(t, "Deque[int]{}", fmt.Sprintf("%#v", d))
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)
	AssertEq(t, "Deque[int]{1, 2, 3}", fmt.Sprintf("%#v", d))
}

func TestEquals(t *testing.T) {
	eq := func(a, b int) bool { return a == b }
	tests := []struct {
		name string
		a    *Deque[int]
		b    *Deque[int]
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

func TestEqualsAfterGrowing(t *testing.T) {
	tests := []int{
		0, 15, 16, 17,
		1023, 1024, 1025,
	}
	eq := func(a, b int) bool {
		return a == b
	}
	for _, n := range tests {
		t.Run(fmt.Sprintf("eq %d", n), func(t *testing.T) {
			d := WithCfg[int](defCfg)
			d.EnsureBack(n)
			slice := make([]int, n)
			for i := range n {
				d.PushBack(i)
				slice[i] = i
			}
			d2 := FromSlice(slice)
			AssertTrue(t, d.Equals(d2, eq))
		})
	}
}

func TestEmptyOperations(t *testing.T) {
	d := WithCfg[int](defCfg)

	val, ok := d.PopFront()
	AssertFalse(t, ok, "PopFront on empty")
	AssertEq(t, 0, val)

	val, ok = d.PopBack()
	AssertFalse(t, ok, "PopBack on empty")
	AssertEq(t, 0, val)

	val, ok = d.Front()
	AssertZeroFalse(t, val, ok, "Front on empty")

	val, ok = d.Back()
	AssertZeroFalse(t, val, ok, "Back on empty")

	val, ok = d.Get(0)
	AssertZeroFalse(t, val, ok, "Get(0) on empty")
}

func TestPushFrontAndPushBack(t *testing.T) {
	d := WithCfg[int](defCfg)
	d.PushBack(2)
	d.PushFront(1)
	d.PushBack(3)
	front, _ := d.Front()
	back, _ := d.Back()
	AssertEq(t, 1, front)
	AssertEq(t, 3, back)
	AssertEq(t, 3, d.Len())
}

func TestPushFront1M(t *testing.T) {
	d := New[int]()
	n := 1_000_000
	want := make([]int, 0, n)
	for i := range n {
		d.PushFront(i)
		want = append(want, i)
	}
	want = ReversedSlice(want)
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestPooledPushFront1M(t *testing.T) {
	d := WithCfg[int](defCfgPooled)
	n := 1_000_000
	want := make([]int, 0, n)
	for i := range n {
		d.PushFront(i)
		want = append(want, i)
	}
	want = ReversedSlice(want)
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestReserveFrontPushFront1M(t *testing.T) {
	d := New[int]()
	n := 1_000_000
	d.ReserveFront(n)
	want := make([]int, 0, n)
	for i := range n {
		d.PushFront(i)
		want = append(want, i)
	}
	want = ReversedSlice(want)
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestPushBack1M(t *testing.T) {
	d := New[int]()
	n := 1_000_000
	want := make([]int, 0, n)
	for i := range n {
		d.PushBack(i)
		want = append(want, i)
	}
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestPooledPushBack1M(t *testing.T) {
	d := WithCfg[int](defCfgPooled)
	n := 1_000_000
	want := make([]int, 0, n)
	for i := range n {
		d.PushBack(i)
		want = append(want, i)
	}
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestReserveBackPushBack1M(t *testing.T) {
	d := New[int]()
	n := 1_000_000
	d.ReserveBack(n)
	want := make([]int, 0, n)
	for i := range n {
		d.PushBack(i)
		want = append(want, i)
	}
	got := d.ToSlice()
	AssertSliceEq(t, want, got)
}

func TestPushBackPopBackPushFront(t *testing.T) {
	d := WithCfg[int](defCfgPooled)
	n := 2025
	for i := 0; i < n/5*2; i++ {
		d.PushBack(i)
	}
	for i := 0; i < n/5; i++ {
		d.PopBack()
	}
	for i := range n {
		d.PushFront(i)
	}
	got := d.ToSlice()
	expected := make([]int, 0, n+n/5)
	for i := n - 1; i >= 0; i-- {
		expected = append(expected, i)
	}
	for i := 0; i < n/5; i++ {
		expected = append(expected, i)
	}
	AssertSliceEq(t, expected, got)
}

func TestPushBackPopBack(t *testing.T) {
	d := WithCfg[int](defCfg)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushBack(v)
	}
	for i := len(slice) - 1; i >= 0; i-- {
		got, ok := d.PopBack()
		AssertEqOk(t, slice[i], got, ok, fmt.Sprintf("PopBack mismatch at i=%d", i))
	}
	AssertEq(t, 0, d.Len())
}

func TestPushFrontPopFront(t *testing.T) {
	d := WithCfg[int](defCfg)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushFront(v)
	}
	for i := len(slice) - 1; i >= 0; i-- {
		got, ok := d.PopFront()
		AssertEqOk(t, slice[i], got, ok, fmt.Sprintf("PopFront mismatch at i=%d", i))
	}
	AssertEq(t, 0, d.Len())
}

func TestPushBackPopFront1025(t *testing.T) {
	d := WithCfg[int](defCfg)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushBack(v)
	}
	for i, v := range slice {
		got, ok := d.PopFront()
		AssertEqOk(t, v, got, ok, fmt.Sprintf("mismatch at i=%d", i))
	}
}

func TestPushFrontPopBack1025(t *testing.T) {
	d := WithCfg[int](defCfg)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushFront(v)
	}
	for i, v := range slice {
		got, ok := d.PopBack()
		AssertEqOk(t, v, got, ok, fmt.Sprintf("mismatch at i=%d", i))
	}
}

func TestNewEnsureBackThenPushBackPopFront1025(t *testing.T) {
	d := New[int]()
	d.EnsureBack(1025)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushBack(v)
	}
	for i, v := range slice {
		got, ok := d.PopFront()
		AssertEqOk(t, v, got, ok, fmt.Sprintf("mismatch at i=%d", i))
	}
}

func TestNewEnsureFrontEnsureBackThenPushBackPopFront1025(t *testing.T) {
	d := New[int]()
	d.EnsureFront(1025) // unused, but needed for testing the stability of such scenario
	d.EnsureBack(1025)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushBack(v)
	}
	for i, v := range slice {
		got, ok := d.PopFront()
		AssertEqOk(t, v, got, ok, fmt.Sprintf("mismatch at i=%d", i))
	}
}

func TestNewEnsureBackEnsureFrontThenPushFrontPopBack1025(t *testing.T) {
	d := New[int]()
	d.EnsureBack(1025) // unused, but needed for testing the stability of such scenario
	d.EnsureFront(1025)
	slice := SliceFromRangeIncl(t, 1, 1025)
	for _, v := range slice {
		d.PushFront(v)
	}
	for i, v := range slice {
		got, ok := d.PopBack()
		AssertEqOk(t, v, got, ok, fmt.Sprintf("mismatch at i=%d", i))
	}
}

func TestGetSmall(t *testing.T) {
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
			d := FromSlice(tt.values)
			gotVal, gotOK := d.Get(tt.idx)
			AssertCommaOk(t, tt.wantVal, tt.wantOK, gotVal, gotOK, fmt.Sprintf("Get(%d)", tt.idx))
		})
	}
}

func TestGetLarge(t *testing.T) {
	// Inclusive range [0..2048]
	slice := SliceFromRangeIncl(t, 0, 2048)
	d := FromSliceCfg(slice, defCfg)
	tests := []int{
		0, 15, 16, 17,
		255, 256, 257,
		1023, 1024, 1025,
		2048,
	}
	for _, idx := range tests {
		t.Run(fmt.Sprintf("index_%d", idx), func(t *testing.T) {
			got, ok := d.Get(idx)
			want := slice[idx]
			AssertEqOk(t, want, got, ok)
		})
	}
}

func TestReserveNOOP(t *testing.T) {
	smallSlice := SliceFromRangeExcl(t, 1, 10)
	largeSlice := SliceFromRangeExcl(t, 1, SuggestBlockSize[int]()*defSideCapBlocks+5)
	tests := []struct {
		name       string
		d          *Deque[int]
		frontItems int
		backItems  int
	}{
		{"empty, 0 0", New[int](), 0, 0},
		{"non-empty, 0 0", FromSlice(smallSlice), 0, 0},
		{"large, 0 0", FromSlice(largeSlice), 0, 0},

		{"empty, 0 -1", New[int](), 0, -1},
		{"non-empty, 0 -1", FromSlice(smallSlice), 0, -1},
		{"large, 0 -1", FromSlice(largeSlice), 0, -1},

		{"empty, -1 0", New[int](), -1, 0},
		{"non-empty, -1 0", FromSlice(smallSlice), -1, 0},
		{"large, -1 0", FromSlice(largeSlice), -1, 0},

		{"empty, -1 -1", New[int](), -1, -1},
		{"non-empty, -1 -1", FromSlice(smallSlice), -1, -1},
		{"large, -1 -1", FromSlice(largeSlice), -1, -1},

		{"empty, 0 5", New[int](), 0, 5},
		{"non-empty, 0 5", FromSlice(smallSlice), 0, 5},
		{"large, 0 5", FromSlice(largeSlice), 0, 5},

		{"empty, 5 0", New[int](), 5, 0},
		{"non-empty, 5 0", FromSlice(smallSlice), 5, 0},
		{"large, 5 0", FromSlice(largeSlice), 5, 0},

		{"empty, -1 500", New[int](), -1, 500},
		{"non-empty, -1 500", FromSlice(smallSlice), -1, 500},
		{"large, -1 500", FromSlice(largeSlice), -1, 500},

		{"empty, 500 -1", New[int](), 500, -1},
		{"non-empty, 500 -1", FromSlice(smallSlice), 500, -1},
		{"large, 500 -1", FromSlice(largeSlice), 500, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.d
			oldMap := d.m
			oldLen := d.Len()
			d.Reserve(tt.frontItems, tt.backItems)
			AssertEq(t, oldLen, d.Len())
			AssertSameSlice(t, oldMap, d.m)
		})
	}
}

func TestReserveFront(t *testing.T) {
	d := New[int]()
	blockSize := d.blockSize

	initMap := d.m

	// determine the number of items to trigger growth
	initialCap := cap(d.m)
	var itemsToGrow int
	for {
		oldCap := cap(d.m)
		d.PushFront(0)
		if cap(d.m) != oldCap {
			break
		}
		itemsToGrow++
	}
	AssertNotSameSlice(t, initMap, d.m)

	// reset deque
	d = New[int]()
	initMap = d.m

	// push exactly itemsToGrow, map should not grow
	for i := range itemsToGrow {
		d.PushFront(i)
	}
	AssertEqual(t, initialCap, cap(d.m))
	AssertSameSlice(t, initMap, d.m)

	// save old blocks
	usedBlocks := d.back.blk - d.front.blk + 1
	oldBlocks := make([][]int, usedBlocks)
	for i := range usedBlocks {
		oldBlocks[i] = d.m[d.front.blk+i]
	}

	// reserve itemsToGrow + 1 and verify growth with reuse
	d.Reserve(itemsToGrow+1, 0)

	neededFrontBlocks := (itemsToGrow + 1 + blockSize - 1) / blockSize
	AssertTrue(t, cap(d.m) >= neededFrontBlocks+usedBlocks)
	AssertEq(t, usedBlocks, d.back.blk-d.front.blk+1)
	for i := range usedBlocks {
		AssertSameSlice(t, oldBlocks[i], d.m[d.front.blk+i])
	}
}

func TestReserveBack(t *testing.T) {
	d := New[int]()
	blockSize := d.blockSize

	initMap := d.m

	// determine the number of items to trigger growth
	initialCap := cap(d.m)
	var itemsToGrow int
	for {
		oldCap := cap(d.m)
		d.PushBack(0)
		if cap(d.m) != oldCap {
			break
		}
		itemsToGrow++
	}
	AssertNotSameSlice(t, initMap, d.m)

	// reset deque
	d = New[int]()
	initMap = d.m

	// push exactly itemsToGrow, map should not grow
	for i := range itemsToGrow {
		d.PushBack(i)
	}
	AssertEqual(t, initialCap, cap(d.m))
	AssertSameSlice(t, initMap, d.m)

	// save old blocks
	usedBlocks := d.back.blk - d.front.blk + 1
	oldBlocks := make([][]int, usedBlocks)
	for i := range usedBlocks {
		oldBlocks[i] = d.m[d.front.blk+i]
	}

	// reserve itemsToGrow + 1 and verify growth with reuse
	d.Reserve(0, itemsToGrow+1)

	neededBackBlocks := (itemsToGrow + 1 + blockSize - 1) / blockSize
	AssertTrue(t, cap(d.m) >= neededBackBlocks+usedBlocks)
	AssertEq(t, usedBlocks, d.back.blk-d.front.blk+1)

	for i := range usedBlocks {
		AssertSameSlice(t, oldBlocks[i], d.m[d.front.blk+i])
	}
}

func TestReserveFrontAndBack(t *testing.T) {
	d := New[int]()
	blockSize := d.blockSize

	initMap := d.m

	// determine the number of items to trigger growth
	initialCap := cap(d.m)
	var itemsToGrow int
	for {
		d.PushBack(0)
		if cap(d.m) != initialCap {
			break
		}
		itemsToGrow++
	}
	AssertNotSameSlice(t, initMap, d.m)

	// reset deque
	d = New[int]()
	initMap = d.m

	// fill at both ends (half front, half back)
	// TODO: can be improved - it's not an API guarantee, but it will do for now
	half := itemsToGrow / 2
	for i := range half {
		d.PushFront(i)
		d.PushBack(i)
	}
	AssertSameSlice(t, initMap, d.m)

	// save old blocks
	usedBlocks := d.back.blk - d.front.blk + 1
	oldBlocks := make([][]int, usedBlocks)
	for i := range usedBlocks {
		oldBlocks[i] = d.m[d.front.blk+i]
	}

	// now reserve front+back growth
	frontItems := half + 1
	backItems := half + 1
	d.Reserve(frontItems, backItems)

	neededFrontBlocks := (frontItems + blockSize - 1) / blockSize
	neededBackBlocks := (backItems + blockSize - 1) / blockSize
	AssertTrue(t, cap(d.m) >= neededFrontBlocks+neededBackBlocks+usedBlocks)
	AssertEq(t, usedBlocks, d.back.blk-d.front.blk+1)

	for i := range usedBlocks {
		AssertSameSlice(t, oldBlocks[i], d.m[d.front.blk+i])
	}
}

func TestEnsureFrontAllocatesAndPreservesBackSlack(t *testing.T) {
	d := New[int]()
	blockSize := d.blockSize

	for i := 0; i < blockSize+3; i++ {
		d.PushBack(i)
	}

	oldMap := d.m
	oldFrontBlk := d.front.blk
	oldBackBlk := d.back.blk

	oldBackSlack := cap(oldMap) - oldBackBlk - 1

	usedBlocks := oldBackBlk - oldFrontBlk + 1
	oldBlocks := make([][]int, usedBlocks)
	for i := range usedBlocks {
		oldBlocks[i] = oldMap[oldFrontBlk+i]
	}

	// Count already allocated blocks before the front
	allocatedBefore := 0
	for i := oldFrontBlk - 1; i >= 0; i-- {
		if oldMap[i] != nil {
			allocatedBefore++
		}
	}

	// Request space larger than existing front capacity to force growth
	items := blockSize*3 + 5

	d.EnsureFront(items)

	// Verify back slack preserved
	newBackSlack := cap(d.m) - d.back.blk - 1
	AssertEq(t, oldBackSlack, newBackSlack)

	// Verify used blocks copied, not reallocated
	newFrontBlk := d.front.blk
	newBackBlk := d.back.blk
	newUsedBlocks := newBackBlk - newFrontBlk + 1

	AssertEq(t, usedBlocks, newUsedBlocks)

	for i := range usedBlocks {
		AssertSameSlice(t, oldBlocks[i], d.m[newFrontBlk+i])
	}

	// Verify required front blocks allocated
	requiredBlocks := (items + blockSize - 1) / blockSize

	remaining := items
	for i := newFrontBlk - 1; remaining > 0 && i >= 0; i-- {
		AssertNotNil(t, d.m[i])
		remaining -= blockSize
	}

	// Verify blocks that were nil before got allocated
	for i := newFrontBlk - requiredBlocks; i < newFrontBlk; i++ {
		if i >= 0 {
			AssertNotNil(t, d.m[i])
		}
	}

	// Verify old map not reused when growth occurred
	if cap(oldMap) != cap(d.m) {
		AssertNotSameSlice(t, oldMap, d.m)
	}
}

// Tests that there is strictly zero allocations after EnsureFront(n)
// and before PushFront(n+k) items where k > 0.
func TestEnsureFrontZeroAllocGuarantee(t *testing.T) {
	tests := []int{
		0,
		1,
		defSideCapBlocks * SuggestBlockSize[int](),
		defSideCapBlocks * SuggestBlockSize[int]() * 2,
		100000,
	}
	for _, n := range tests {
		t.Run(fmt.Sprintf("EnsureFront(%d)", n), func(t *testing.T) {
			d := New[int]()
			d.EnsureFront(n)

			oldMap := d.m
			oldCap := cap(d.m)
			oldEntries := make([][]int, cap(oldMap))
			copy(oldEntries, oldMap)

			for i := range n {
				d.PushFront(i)
			}

			AssertSameSlice(t, oldMap, d.m, "Map must not be replaced")
			AssertEq(t, oldCap, cap(d.m), "Map must have the same capacity")
			AssertEq(t, len(oldEntries), len(d.m), "All blocks should be the same")
			for i := range oldEntries {
				AssertSameSlice(t, oldEntries[i], d.m[i], fmt.Sprintf("mismatch at index %d", i))
			}
		})
	}
}

func TestGetPtrMutation(t *testing.T) {
	d := FromSlice([]int{1, 2, 3})
	ptr, ok := d.GetPtr(1)
	AssertTrue(t, ok)
	*ptr = 99
	val, ok := d.Get(1)
	AssertEqOk(t, 99, val, ok)
}

func TestPushBackGrowthStability(t *testing.T) {
	d := WithCfg[int](defCfg)
	stages := []struct {
		name string
		from int
		to   int
	}{
		{"push16", 0, 15},
		{"push100", 16, 115},
		{"push1000", 116, 1115},
	}
	for _, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			for i := stage.from; i <= stage.to; i++ {
				d.PushBack(i)
			}
			AssertDequeInvariant(t, d)
			for i := 0; i < d.Len(); i++ {
				got, ok := d.Get(i)
				want := i
				AssertEqOk(t, want, got, ok, fmt.Sprintf("stage %s: Get(%d)", stage.name, i))
				if want != got {
					t.FailNow()
				}
			}
		})
	}
}

func TestPushFrontGrowthStability(t *testing.T) {
	d := WithCfg[int](defCfg)
	stages := []struct {
		name string
		from int
		to   int
	}{
		{"push16", 0, 15},
		{"push100", 16, 115},
		{"push1000", 116, 1115},
	}
	for _, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			for i := stage.from; i <= stage.to; i++ {
				d.PushFront(i)
			}
			AssertDequeInvariant(t, d)
			for i := d.Len() - 1; i >= 0; i-- {
				got, ok := d.Get(i)
				want := d.Len() - i - 1
				AssertEqOk(t, want, got, ok, fmt.Sprintf("stage %s: Get(%d)", stage.name, i))
			}
		})
	}
}

func TestRandomOps(t *testing.T) {
	d := WithCfg[int](defCfg)
	ref := []int{}
	rnd := rand.New(rand.NewSource(1))
	for range 100000 {
		op := rnd.Intn(4)
		switch op {
		case 0: // push back
			v := rnd.Int()
			d.PushBack(v)
			ref = append(ref, v)
		case 1: // push front
			v := rnd.Int()
			d.PushFront(v)
			ref = append([]int{v}, ref...)
		case 2: // pop front
			if len(ref) == 0 {
				continue
			}
			want := ref[0]
			ref = ref[1:]
			got, ok := d.PopFront()
			AssertEqOk(t, want, got, ok)
		case 3: // pop back
			if len(ref) == 0 {
				continue
			}
			want := ref[len(ref)-1]
			ref = ref[:len(ref)-1]
			got, ok := d.PopBack()
			AssertEqOk(t, want, got, ok)
		}
		AssertDequeInvariant(t, d)
		AssertSliceEq(t, ref, d.ToSlice())
	}
}

func TestBlockBoundaryWrap(t *testing.T) {
	d := WithCfg[int](defCfg)
	for i := range 100 {
		d.PushBack(i)
	}
	AssertDequeInvariant(t, d)
	for range 90 {
		d.PopFront()
	}
	AssertDequeInvariant(t, d)
	for i := 100; i < 200; i++ {
		d.PushBack(i)
	}
	AssertDequeInvariant(t, d)
	AssertSliceEq(t, SliceFromRangeIncl(t, 90, 199), d.ToSlice())
}

func TestDequeBlockBoundary(t *testing.T) {
	cfg := defCfg
	cfg.BlockSize = 8
	d := WithCfg[int](cfg)
	for i := range 1000 {
		d.PushBack(i)
	}
	for i := range 1000 {
		v, _ := d.PopFront()
		AssertEq(t, i, v)
	}
}

func TestShrinkToFit(t *testing.T) {
	tests := []struct {
		name      string
		initCap   int
		pushCount int
		pooled    bool
	}{
		{"empty", 4, 0, false},
		{"single block", 4, 2, false},
		{"multiple blocks", 4, 10, false},
		{"exactly initCap", 4, 16, false},
		{"pooled small", 4, 2, true},
		{"pooled large", 4, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defCfg
			cfg.BlockSize = 4
			cfg.Pooled = tt.pooled
			d := WithCfg[int](cfg)
			for i := 0; i < tt.pushCount; i++ {
				d.PushBack(i)
			}

			d.ShrinkToFit()

			oldLen := d.Len()
			AssertEq(t, oldLen, d.Len(), "length should be preserved")
			AssertTrue(t, d.front.blk >= d.initCfg.frontBlock, "front slack should be preserved")
			AssertTrue(t, d.back.blk < cap(d.m), "back slack should be preserved")
			AssertTrue(t, cap(d.m) >= d.initCfg.totalBlocks,
				"capacity should never be less than initConfig total blocks")
			for i := 0; i < d.Len(); i++ {
				got, ok := d.Get(i)
				AssertEqOk(t, i, got, ok)
			}

			d.PushBack(tt.pushCount)
			got, ok := d.Back()
			AssertEqOk(t, tt.pushCount, got, ok, "push after shrink should work")
		})
	}
}

func TestClear(t *testing.T) {
	d := WithCfg[int](defCfg)
	slice := SliceFromRangeIncl(t, 1, 128)
	for _, v := range slice {
		d.PushBack(v)
	}
	AssertDequeInvariant(t, d)
	d.Clear()
	AssertDequeInvariant(t, d)
	AssertEq(t, 0, d.Len())
	AssertSliceEq(t, []int{}, d.ToSlice())
}

func TestClearThenReuse(t *testing.T) {
	d := WithCfg[int](defCfg)
	for i := range 256 {
		d.PushBack(i)
	}
	AssertDequeInvariant(t, d)
	d.Clear()
	AssertDequeInvariant(t, d)
	for i := range 128 {
		d.PushBack(i)
	}
	AssertDequeInvariant(t, d)
	got := d.ToSlice()
	want := SliceFromRangeIncl(t, 0, 127)
	AssertSliceEq(t, want, got)
}

func TestClearRelease(t *testing.T) {
	cfgs := []DequeCfg{
		defCfg,
		defCfgPooled,
	}

	for _, cfg := range cfgs {
		t.Run(fmt.Sprintf("pooled=%v", cfg.Pooled), func(t *testing.T) {
			d := WithCfg[int](cfg)
			for i := range 5000 {
				d.PushBack(i)
			}
			for range 2000 {
				d.PopFront()
			}
			for i := 5000; i < 8000; i++ {
				d.PushFront(i)
			}
			// ensure expanded
			AssertTrue(t, len(d.m) > d.initCfg.totalBlocks)

			d.ClearRelease()
			AssertEq(t, 0, d.Len())
			AssertEq(t, d.initCfg.totalBlocks, len(d.m))
			AssertEq(t, d.initCfg.frontBlock, d.front.blk)
			AssertEq(t, d.initCfg.backBlock, d.back.blk)

			AssertEq(t, 0, d.front.off)
			AssertEq(t, 0, d.back.off)

			for i := range 100 {
				d.PushBack(i)
			}
			got := d.ToSlice()
			want := SliceFromRangeIncl(t, 0, 99)
			AssertSliceEq(t, want, got)

			AssertDequeInvariant(t, d)
		})
	}
}

func TestClearReleaseRepeated(t *testing.T) {
	d := WithCfg[int](defCfgPooled)
	for range 10 {
		for i := range 1000 {
			d.PushBack(i)
		}
		for range 500 {
			d.PopFront()
		}

		d.ClearRelease()

		AssertEq(t, 0, d.Len())
		AssertEq(t, d.initCfg.totalBlocks, len(d.m))

		AssertDequeInvariant(t, d)
	}
}

func TestClearReleasePushBackHangOrPanic(t *testing.T) {
	cfg := defCfg
	d := WithCfg[int](cfg)
	const maxItems = 31
	for n := 1; n <= maxItems; n++ {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			d.ClearRelease()
			for i := 0; i < n; i++ {
				d.PushBack(i)
			}
			d.ClearRelease()
			d.PushBack(0)
			d.PushBack(1)
		})
	}
}

func TestNastyInterleave1(t *testing.T) {
	d := New[int]()
	ref := []int{}
	for i := range 10000 {
		switch i % 4 {
		case 0:
			d.PushFront(i)
			ref = append([]int{i}, ref...)
		case 1:
			d.PushBack(i)
			ref = append(ref, i)
		case 2:
			if len(ref) > 0 {
				v1, _ := d.PopFront()
				v2 := ref[0]
				ref = ref[1:]
				AssertEq(t, v2, v1)
			}
		case 3:
			if len(ref) > 0 {
				v1, _ := d.PopBack()
				v2 := ref[len(ref)-1]
				ref = ref[:len(ref)-1]
				AssertEq(t, v2, v1)
			}
		}
		AssertSliceEq(t, ref, d.ToSlice())
	}
}

func TestNastyInterleave2(t *testing.T) {
	const n = 1024
	d := WithCfg[int](defCfgPooled)

	for i := range n {
		if i%2 == 0 {
			d.PushBack(i)
		} else {
			d.PushFront(-i)
		}
	}
	AssertDequeInvariant(t, d)

	for range n / 4 {
		d.PopFront()
		d.PopBack()
	}
	AssertDequeInvariant(t, d)

	// push again to force wrapping around block boundaries
	for i := n; i < n+n/2; i++ {
		if i%3 == 0 {
			d.PushFront(i)
		} else {
			d.PushBack(-i)
		}
	}
	AssertDequeInvariant(t, d)

	oldCap := cap(d.m)
	d.ShrinkToFit()
	AssertDequeInvariant(t, d)

	// capacity should have decreased but not below initCfg.totalBlocks
	AssertTrue(t, cap(d.m) <= oldCap, "ShrinkToFit did not reduce capacity")
	AssertTrue(t, cap(d.m) >= d.initCfg.totalBlocks, "ShrinkToFit shrank below init config")
	AssertTrue(t, d.front.blk >= 0 && d.back.blk < cap(d.m), "front/back indices in bounds")

	// clear and then clear-release
	d.Clear()
	AssertDequeInvariant(t, d)
	AssertEq(t, 0, d.Len())

	d.ClearRelease()
	AssertDequeInvariant(t, d)
	AssertEq(t, 0, d.Len())
	AssertTrue(t, cap(d.m) >= d.initCfg.totalBlocks, "ClearRelease should allocate at least init blocks")

	// push/pop after shrink+clear-release
	for i := 0; i < 10*d.initCfg.totalBlocks; i++ {
		if i%2 == 0 {
			d.PushBack(i)
		} else {
			d.PushFront(-i)
		}
	}
	AssertDequeInvariant(t, d)
	items := d.ToSlice()
	AssertEq(t, d.Len(), len(items))

	// randomly interleave pops to force block reclaim/pointer juggling
	for i := 0; i < len(items)/2; i++ {
		if i%2 == 0 {
			_, _ = d.PopFront()
		} else {
			_, _ = d.PopBack()
		}
	}
	AssertDequeInvariant(t, d)
}

func TestSliceAndPtrSlice(t *testing.T) {
	even := SliceFromRangeIncl(t, 1, 6)
	odd := even[0 : len(even)-1]
	tests := []struct {
		name  string
		slice []int
		start int
		end   int
	}{
		{"empty slice from empty list", []int{}, 0, 3},
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
			var want []int
			if start >= 0 && start < len(slice) && tt.end > start {
				want = slice[start:min(tt.end, len(slice))]
			} else {
				want = []int{}
			}

			l := FromSlice(slice)
			got := l.Slice(tt.start, tt.end)
			AssertSliceEq(t, want, got, "Slice")
			ptrGot := l.PtrSlice(tt.start, tt.end)
			AssertPtrSliceEq(t, want, ptrGot, "PtrSlice")
		})
	}
}

func TestBidiIterNavigation(t *testing.T) {
	cases := [][]int{
		SliceFromRangeIncl(t, 1, 3),
		SliceFromRangeIncl(t, 1, defCfg.BlockSize),
		SliceFromRangeIncl(t, 1, defCfg.BlockSize+1),
		SliceFromRangeIncl(t, 1, defCfgInitCap),
		SliceFromRangeIncl(t, 1, defCfgInitCap+1),
		SliceFromRangeIncl(t, 1, defCfgInitCap*2),
	}

	for _, slice := range cases {
		d := FromSlice(slice)
		it := d.BidiIter()

		AssertTrue(t, it.HasNext(), "HasNext at start should be true")
		AssertFalse(t, it.HasPrev(), "HasPrev at start should be false")

		val, ok := it.Peek()
		AssertEqOk(t, slice[0], val, ok, "Peek should show first item without advancing")
		val, ok = it.Next()
		AssertEqOk(t, slice[0], val, ok, "Next should advance into first item")
		AssertTrue(t, it.HasPrev(), "HasPrev after first next should be true")
		val, ok = it.PeekBack()
		AssertEqOk(t, slice[0], val, ok, "PeekBack should give the item we just traversed")

		val, ok = it.Next()
		AssertEqOk(t, slice[1], val, ok, "Next should advance further")
		val, ok = it.Prev()
		AssertEqOk(t, slice[1], val, ok, "Prev should step back")
		val, ok = it.Peek()
		AssertEqOk(t, slice[1], val, ok, "Peek after Prev should show second")

		got := it.Reset().TakeSlice(len(slice))
		AssertSliceEq(t, slice, got)
	}
}

func TestIterTakeWhileAndSkip(t *testing.T) {
	d := FromSlice(SliceFromRangeIncl(t, 1, 5))

	// TakeWhile should stop at first non-match
	it := d.BidiIter()
	got := it.TakeWhile(func(v int) bool { return v < 3 })
	AssertSliceEq(t, []int{1, 2}, got, "TakeWhile")

	// TakeWhilePtr should stop at first non-match
	it = d.BidiIter()
	ptrs := it.TakeWhilePtr(func(p *int) bool { return *p < 4 })
	AssertPtrSliceEq(t, []int{1, 2, 3}, ptrs, "TakeWhilePtr")

	// Skip should advance by n
	it = d.BidiIter()
	it.Skip(2)
	val, ok := it.Next()
	AssertEqOk(t, 3, val, ok, "Skip(2) then Next")

	// SkipBack should step back by n
	it.ResetBack()
	it.SkipBack(2)
	val, ok = it.Prev()
	AssertEqOk(t, 3, val, ok, "SkipBack(2) then Prev")

	// SkipWhile should advance until predicate fails
	it = d.BidiIter()
	it.SkipWhile(func(v int) bool { return v < 4 })
	val, ok = it.Next()
	AssertEqOk(t, 4, val, ok, "SkipWhile then Next")

	// SkipWhilePtr should advance until predicate fails
	it = d.BidiIter()
	it.SkipWhilePtr(func(p *int) bool { return *p < 5 })
	val, ok = it.Next()
	AssertEqOk(t, 5, val, ok, "SkipWhilePtr then Next")
}

func TestIterReset(t *testing.T) {
	d := FromSlice(SliceFromRangeIncl(t, 1, defCfgInitCap))
	{
		it := d.BidiIter().Skip(5)
		AssertTrue(t, it.HasPrev(), "BidiIter has next after skip (sanity)")
		AssertTrue(t, it.HasNext(), "BidiIter has next after skip (sanity)")
		it.Reset()
		AssertFalse(t, it.HasPrev(), "BidiIter has no prev after Reset")
		AssertTrue(t, it.HasNext(), "BidiIter has next after Reset")
		it.ResetBack()
		AssertTrue(t, it.HasPrev(), "BidiIter has rev after ResetBack")
		AssertFalse(t, it.HasNext(), "BidiIter has no next after Reset")
	}
	{
		it := d.Iter().Skip(5)
		AssertTrue(t, it.HasNext(), "Iter has next after skip (sanity)")
		it.Reset()
		AssertTrue(t, it.HasNext(), "Iter has next after Reset")
		front, _ := d.PopFront()
		firstAfterReset, _ := it.Next()
		AssertEq(t, front, firstAfterReset, "Iter next after Reset returns front")
	}
	{
		it := d.RevIter().Skip(5)
		AssertTrue(t, it.HasNext(), "RevIter has next after skip (sanity)")
		it.Reset()
		AssertTrue(t, it.HasNext(), "RevIter has next after Reset")
		back, _ := d.PopBack()
		firstAfterReset, _ := it.Next()
		AssertEq(t, back, firstAfterReset, "RevIter next after Reset returns back")
	}
}

func TestIterSkipLargeResetsBack(t *testing.T) {
	d := FromSlice(SliceFromRangeIncl(t, 1, defCfgInitCap))
	it := d.BidiIter().Skip(d.Len() + 1)
	AssertTrue(t, it.HasPrev(), "skip larger than rem length should have prev")
	AssertFalse(t, it.HasNext(), "skip larger than rem length should reset back have no next")
	back, _ := d.Back()
	backFromIter, _ := it.Prev()
	AssertEq(t, back, backFromIter)
	d.PushBack(100) // force grow
	it = d.BidiIter().Skip(d.Len() + 1)
	AssertTrue(t, it.HasPrev(), "skip larger than rem length after push should have prev")
	AssertFalse(t, it.HasNext(), "skip larger than rem length after push should have no next")
	backFromIter, _ = it.Prev()
	AssertEq(t, 100, backFromIter)
	it = d.BidiIter().Skip(d.Len() * 100)
	AssertTrue(t, it.HasPrev(), "skip larger than rem length should have prev")
	AssertFalse(t, it.HasNext(), "skip larger than rem length should reset back have no next")
	backFromIter, _ = it.Prev()
	AssertEq(t, 100, backFromIter)
}

func TestIterSkipBackLargeResetsFront(t *testing.T) {
	d := FromSlice(SliceFromRangeIncl(t, 1, defCfgInitCap))
	it := d.BidiIter().SkipBack(1)
	AssertFalse(t, it.HasPrev(), "SkipBack 1 at start should have no prev")
	AssertTrue(t, it.HasNext(), "SkipBack 1 at start should have next")
	front, _ := d.Front()
	frontFromIter, _ := it.Next()
	AssertEq(t, front, frontFromIter)
	d.PushBack(100) // force grow
	it = d.BidiIter().Skip(d.Len() / 2).SkipBack(d.Len() * 100)
	AssertFalse(t, it.HasPrev(), "skip larger than length should have no prev")
	AssertTrue(t, it.HasNext(), "skip larger than length should reset back have next")
	frontFromIter, _ = it.Next()
	AssertEq(t, 1, frontFromIter)
}

func TestIterTakeAndSkip(t *testing.T) {
	n := 5
	items := SliceFromRangeIncl(t, 1, n)
	revItems := ReversedSlice(items)
	l := FromSlice(items)
	tests := []struct {
		name  string
		it    DequeIterator[int]
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

// ----- Helpers -----
func AssertDequeInvariant[T any](t *testing.T, d *Deque[T]) {
	t.Helper()

	dLen := d.len
	mapLen := len(d.m)

	AssertTrue(t, dLen >= 0, "deque len must be nonneg")
	AssertFalse(t, dLen != 0 && mapLen == 0, fmt.Sprintf(
		"deque has zero blocks but non-zero length: %d", dLen))

	front := d.front
	back := d.back
	blockSize := d.blockSize

	if d.len == 0 {
		AssertEq(t, front.blk, back.blk)
		AssertEq(t, front.off, back.off)
	}

	AssertTrue(t, front.off >= 0, "front offset must be nonneg")
	AssertTrue(t, front.off < blockSize, "front offset must fit block size")

	AssertTrue(t, back.off >= 0, "back offset must be nonneg")
	AssertTrue(t, back.off < blockSize, "back offset must fit block size")

	AssertTrue(t, front.blk >= 0, "front blk must be nonneg")
	AssertTrue(t, back.blk >= 0, "back blk must be nonneg")

	AssertTrue(t, front.blk < mapLen, "front blk must fit map len")
	AssertTrue(t, back.blk < mapLen, "back blk must fit map len")

	AssertTrue(t, front.blk <= back.blk, "front blk <= back blk")
	AssertFalse(t, front.blk == back.blk && front.off > back.off,
		fmt.Sprintf("front offset > back offset within same block: front=%v back=%v",
			front, back))

	for i := front.blk; i <= back.blk && i < len(d.m); i++ {
		AssertNotNil(t, d.m[i], fmt.Sprintf("block %d is nil but should exist", i))
	}
}

func TestDequeExample1(t *testing.T) {
	d := FromSlice([]int{2, 4, 8, 16})
	v, ok := d.PopBack()
	AssertEqOk(t, 16, v, ok)
	d.PushFront(1)
	if v, ok := d.PopFront(); ok {
		AssertEq(t, 1, v)
	}
	AssertSliceEq(t, []int{2, 4, 8}, d.ToSlice())
	d.Iter().Skip(1).ForEachPtr(func(x *int) bool {
		*x *= 10
		return true
	})
	AssertSliceEq(t, []int{2, 40, 80}, d.ToSlice())
}
