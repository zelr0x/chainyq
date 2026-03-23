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

// Package deque provides [Deque] - a deque based on a segmented array.
package deque

import (
	"fmt"
	"math"
	"math/bits"
	"strings"
	"unsafe"

	"github.com/zelr0x/chainyq"
	"github.com/zelr0x/chainyq/internal/numutil"
	"github.com/zelr0x/chainyq/seq"
)

const (
	defSideCapBlocks int = 8
	minBlockSize     int = 4
	maxBlockSize     int = 4096
	maxSideCapItems  int = math.MaxInt - maxBlockSize + 1 // Overflow protection only
	takeWhileInitCap int = 32
)

type DequeCfg struct {
	// BlockSize is the size of the block in the deque.
	// Must be power of 2. If The passed value is not a power of two,
	// it will be rounded up to the next power of 2.
	// Use [SuggestBlockSize] to get advised block size.
	BlockSize int

	// FrontCap is the number of items you can push to the front
	// without reallocating the deque. Note that this is not the same
	// as preallocating the blocks themselves. For that use [EnsureFront]
	// and [EnsureBack] after creation.
	//
	// You can also reserve space after the deque is creeated with [Reserve]
	// or [ReserveFront], but it's better to specify it here if you know
	// the wanted capacity before deque creation to not risk wasting
	// the initial allocation.
	FrontCap int

	// BackCap is the number of items you can push to the back
	// without reallocating the deque. Note that this is not the same
	// as preallocating the blocks themselves. For that use [EnsureFront]
	// and [EnsureBack] after creation.
	//
	// You can also reserve space after the deque is creeated with [Reserve]
	// or [ReserveBack], but it's better to specify it here if you know
	// the wanted capacity before deque creation to not risk wasting
	// the initial allocation.
	BackCap int

	// Pooled specifies if object pool should be used. See [PreallocPool].
	// Note that using a pool does not automatically improve performance.
	// It just means blocks (underlying slices) are kept in the pool
	// instead of being freed when not in use, and future allocations
	// reuse them rather than allocating new slices.
	//
	// Both this deque and the Go runtime handle occasional slice allocations
	// very well, and it is difficult to beat the runtime allocator even
	// with an optimized pool. Don't treat pooling as a free speed
	// boost - you may see no gain or even a slight slowdown.
	// The main benefit is more predictable memory usage, but the gains
	// dependend heavily on the GC and current load.
	//
	// If you know the number of items you will push and want to maximize speed,
	// use [EnsureFront] or [EnsureBack] instead.
	Pooled bool
}

type initCfg struct {
	totalBlocks int
	frontBlock  int
	backBlock   int
}

type dequeState[T any] struct {
	m     [][]T
	front blockIndex // first item index
	back  blockIndex // index of the slot just after last item
	len   int
}

type blockCfg struct {
	blockSize  int  // immutable
	blockMask  int  // immutable
	blockShift uint // immutable
}

type Deque[T any] struct {
	dequeState[T]               // mutable, 64 bytes on x64
	blockCfg                    // immutable, 24 bytes on x64
	a             dequeAlloc[T] // immutable, internal state, 40 bytes on x64
	initCfg       initCfg       // immutable, 24 bytes on x64
}

// Iter is a bidirectional iterator for [Deque]. The iterator is invalidated
// by all operations that mutate the iterated deque.
type BidiIter[T any] struct {
	d dequeState[T]
	blockCfg
	cur      blockIndex
	frontAbs int // absolute idx of front, immutable
	pos      int // 0-based idx in items since start (not flipped for RevIter)
}

// Iter is a forward iterator for [Deque]. The iterator is invalidated
// by all operations that mutate the iterated deque.
type Iter[T any] struct {
	b BidiIter[T]
}

// Iter is a reverse iterator for a [Deque]. The iterator is invalidated
// by all operations that mutate the iterated deque.
type RevIter[T any] struct {
	b BidiIter[T]
}

type DequeIterator[T any] interface {
	chainyq.Iterator[T]
	chainyq.Sequencer[T]
	ForEach(f func(T) bool)
	ForEachPtr(f func(*T) bool)
	ToChan(size int) <-chan T
	ToPtrChan(size int) <-chan *T
	TakeSlice(n int) []T
	TakePtrSlice(n int) []*T
	TakeWhile(pred func(T) bool) []T
	TakeWhilePtr(pred func(*T) bool) []*T
}

// blockIndex is a position inside the deque.
type blockIndex struct {
	blk int
	off int
}

// New creates a deque with default block size and default capacity.
func New[T any]() *Deque[T] {
	return WithCfg[T](DequeCfg{})
}

func NewPooled[T any]() *Deque[T] {
	return WithCfg[T](DequeCfg{
		Pooled: true,
	})
}

// WithCfg creates a deque with the given configuration.
func WithCfg[T any](cfg DequeCfg) *Deque[T] {
	res := NewValue[T](cfg)
	return &res
}

// NewValue is a deque constructor for cases when you need deque as a value.
// This is meant for advanced use cases, most users should use [New]
// and [WithCfg] to create a *Deque[T].
func NewValue[T any](cfg DequeCfg) Deque[T] {
	blockSize := numutil.ClampInt(cfg.BlockSize, 0, maxBlockSize)
	if blockSize == 0 {
		blockSize = SuggestBlockSize[T]()
	} else {
		blockSize = int(numutil.RoundNextPow2(uint(blockSize))) // #nosec G115
	}

	frontBlocks := defSideCapBlocks
	backBlocks := defSideCapBlocks
	{
		frontCapItems := numutil.ClampInt(cfg.FrontCap, 0, maxSideCapItems)
		backCapItems := numutil.ClampInt(cfg.BackCap, 0, maxSideCapItems)
		if frontCapItems != 0 {
			frontBlocks = numutil.MaxInt(1, blocksForCapCeil(blockSize, frontCapItems))
		}
		if backCapItems != 0 {
			backBlocks = numutil.MaxInt(1, blocksForCapCeil(blockSize, backCapItems))
		}
	}

	d := Deque[T]{
		dequeState: dequeState[T]{},
		blockCfg: blockCfg{
			blockSize:  blockSize,
			blockMask:  blockSize - 1,
			blockShift: uint(bits.TrailingZeros(uint(blockSize))), // #nosec G115
		},
		a: newDequeAlloc[T](dequeAllocCfg{
			blockSize: blockSize,
			pooled:    cfg.Pooled,
		}),
		initCfg: initCfg{
			totalBlocks: numutil.MaxInt(2, frontBlocks+backBlocks),
			frontBlock:  frontBlocks,
			backBlock:   frontBlocks, // not a typo
		},
	}
	d.init()
	return d
}

// SuggestBlockSize suggests a block size that should be good for a given type.
// This can then be used to call [WithCfg] - it is exactly how [New] does it.
func SuggestBlockSize[T any]() int {
	const (
		cacheLine   = 64
		targetBytes = 1024
		minBytes    = cacheLine
		maxBytes    = 4096
	)

	var zero T
	sz := int(unsafe.Sizeof(zero)) // #nosec G115
	if sz == 0 {
		return minBlockSize
	}

	blockBytes := numutil.ClampInt(targetBytes, minBytes, maxBytes)

	blockSize := (blockBytes + sz - 1) / sz
	blockSize = int(numutil.RoundNextPow2(uint(blockSize))) // #nosec G115
	blockSize = numutil.MaxInt(blockSize, minBlockSize)

	bytesUsed := blockSize * sz
	rem := bytesUsed % cacheLine
	if rem != 0 {
		padding := (cacheLine - rem + sz - 1) / sz
		blockSize += padding
		blockSize = int(numutil.RoundNextPow2(uint(blockSize))) // #nosec G115
	}

	return blockSize
}

// FromSlice creates a new deque from the specified slice.
func FromSlice[T any](vals []T) *Deque[T] {
	d := New[T]()
	for _, v := range vals {
		d.PushBack(v)
	}
	return d
}

// FromSliceCfg creates a new deuqe from the specified slice with given config.
func FromSliceCfg[T any](vals []T, cfg DequeCfg) *Deque[T] {
	d := WithCfg[T](cfg)
	for _, v := range vals {
		d.PushBack(v)
	}
	return d
}

// Len returns the length of a deque. Returns 0 for nil.
func Len[T any](d *Deque[T]) int {
	if d == nil {
		return 0
	}
	return d.len
}

// String implements fmt.Stringer, used for %v and %+v.
func (d *Deque[T]) String() string {
	if d == nil {
		return "nil"
	}
	if d.len == 0 {
		return "Deque[]"
	}
	var sb strings.Builder
	sb.Grow(7 + 2*d.len)
	sb.WriteString("Deque[")
	it := d.Iter()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		fmt.Fprintf(&sb, "%v", v)
		if !it.HasNext() {
			break
		}
		sb.WriteString(" ")
	}
	sb.WriteRune(']')
	return sb.String()
}

// GoString implements fmt.GoStringer, used for %#v.
func (d *Deque[T]) GoString() string {
	if d == nil {
		return "nil"
	}
	var zero T
	if d.len == 0 {
		return fmt.Sprintf("Deque[%T]{}", zero)
	}
	var sb strings.Builder
	sb.Grow(32 + 3*d.len)
	fmt.Fprintf(&sb, "Deque[%T]{", zero)
	it := d.Iter()
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		fmt.Fprintf(&sb, "%#v", v)
		if !it.HasNext() {
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteRune('}')
	return sb.String()
}

// Equals returns true if both deques are nil or this deque
// is element-wise equal to the specified deque according to the specified
// equality function, false otherwise. Empty deque is not equal to nil deque.
func (d *Deque[T]) Equals(other *Deque[T], eq func(T, T) bool) bool {
	switch {
	case d == nil:
		return other == nil
	case other == nil:
		return false
	case d.len != other.len:
		return false
	case d.len == 0:
		return true
	}
	itA := d.Iter()
	itB := other.Iter()
	var a, b T
	var okA, okB bool
	for {
		a, okA = itA.Next()
		b, okB = itB.Next()
		if !okA || !okB {
			return !okA && !okB
		}
		if !eq(a, b) {
			return false
		}
	}
}

// Len returns the number of items in the deque.
func (d *Deque[T]) Len() int {
	return d.len
}

// IsEmpty checks if the deque is empty.
func (d *Deque[T]) IsEmpty() bool {
	return d.len == 0
}

// Front returns the item at the front of the deque and true
// if it exists, zero value and false otherwise.
func (d *Deque[T]) Front() (T, bool) {
	if d.len == 0 {
		var zero T
		return zero, false
	}
	return d.m[d.front.blk][d.front.off], true
}

// FrontPtr returns a pointer to the item at the front of the deque and true
// if it exists, nil and false otherwise.
func (d *Deque[T]) FrontPtr() (*T, bool) {
	if d.len == 0 {
		return nil, false
	}
	return &d.m[d.front.blk][d.front.off], true
}

// Back returns the item at the back of the deque and true
// if it exists, zero value and false otherwise.
func (d *Deque[T]) Back() (T, bool) {
	if d.len == 0 {
		var zero T
		return zero, false
	}
	back := d.back
	blk := back.blk
	off := back.off - 1
	if off < 0 {
		blk--
		off = d.blockSize - 1
	}
	return d.m[blk][off], true
}

// BackPtr returns a pointer to the item at the back of the deque and true
// if it exists, zero value and false otherwise.
func (d *Deque[T]) BackPtr() (*T, bool) {
	if d.len == 0 {
		return nil, false
	}
	back := d.back
	blk := back.blk
	off := back.off
	if off == 0 {
		blk--
		off = d.blockSize - 1
	}
	return &d.m[blk][off], true
}

// PushFront adds the specified item to the front of the deque.
func (d *Deque[T]) PushFront(val T) {
	// PushFront adds right before d.front.
	blk := d.front.blk
	off := (d.front.off - 1) & d.blockMask

	if off == d.blockMask {
		if blk == 0 {
			d.growFrontBy(1)
			blk = d.front.blk
		}
		blk--
		if d.m[blk] == nil {
			d.m[blk] = d.a.NewBlock()
		}
	}

	d.m[blk][off] = val
	d.front.blk = blk
	d.front.off = off
	d.len++
}

// PushBack adds the specified item to the back of the deque.
func (d *Deque[T]) PushBack(val T) {
	// PushBack adds right at d.back.
	blk := d.back.blk
	off := d.back.off
	block := d.m[blk]

	block[off] = val
	d.len++

	off = (off + 1) & d.blockMask
	if off == 0 {
		// Don't increment blk and set off to 0 immediately here so that
		// growBackBy correctly calculates the number of used blocks.
		nextBlockIdx := blk + 1
		if nextBlockIdx == len(d.m) {
			d.growBackBy(1)
		} else {
			nextBlock := d.m[nextBlockIdx]
			if nextBlock == nil {
				nextBlock = d.a.NewBlock()
				d.m[nextBlockIdx] = nextBlock
			}
		}
		d.back.blk = nextBlockIdx
	}
	d.back.off = off
}

// PopFront removes and returns the item at the front of the deque
// and true if it exists, zero value and false otherwise.
func (d *Deque[T]) PopFront() (T, bool) {
	// PopFront removes right at d.front.
	if d.len == 0 {
		var zero T
		return zero, false
	}

	blk := d.front.blk
	off := d.front.off
	block := d.m[blk]

	val := block[off]
	d.len--

	off = (off + 1) & d.blockMask
	if off == 0 {
		if d.a.pooled {
			// Reclaim block when this one is unused, always leaving 1 block reserved.
			prevBlkIdx := d.front.blk - 1
			if prevBlkIdx >= 0 && prevBlkIdx < len(d.m) {
				d.reclaimBlock(prevBlkIdx)
			}
		}

		d.front.blk = blk + 1
	}
	d.front.off = off
	return val, true
}

// PopBack removes and returns the item at the back of the deque
// and true if it exists, zero value and false otherwise.
func (d *Deque[T]) PopBack() (T, bool) {
	// PopFront removes right before d.back.
	if d.len == 0 {
		var zero T
		return zero, false
	}

	blk := d.back.blk
	off := (d.back.off - 1) & d.blockMask

	if off == d.blockMask {
		if d.a.pooled {
			// Reclaim block when this one is unused, always leaving 1 block reserved.
			nextBlkIdx := blk + 1
			if nextBlkIdx >= 0 && nextBlkIdx < len(d.m) {
				d.reclaimBlock(nextBlkIdx)
			}
		}
		blk--
		d.back.blk = blk
	}

	block := d.m[blk]
	val := block[off]

	d.back.off = off
	d.len--
	return val, true
}

// Get finds an item with the given index from the deque and returns
// it and true if the index is valid, zero value and false otherwise.
func (d *Deque[T]) Get(idx int) (T, bool) {
	if idx < 0 || idx >= d.len {
		var zero T
		return zero, false
	}
	abs := idx + d.front.off
	blk := d.front.blk + (abs >> d.blockShift)
	off := abs & d.blockMask
	return d.m[blk][off], true
}

// Get finds an item with the given index from the deque and returns
// a pointer to it and true if the index is valid, zero value
// and false otherwise.
func (d *Deque[T]) GetPtr(idx int) (*T, bool) {
	if idx < 0 || idx >= d.len {
		return nil, false
	}
	abs := idx + d.front.off
	blk := d.front.blk + (abs >> d.blockShift)
	off := abs & d.blockMask
	return &d.m[blk][off], true
}

// Set replaces the value at specified index with the specified new value
// and returns old value and true on success, new value and false otherwise.
func (d *Deque[T]) Set(idx int, v T) (T, bool) {
	if idx < 0 || idx >= d.len {
		return v, false
	}
	abs := idx + d.front.off
	blk := d.front.blk + (abs >> d.blockShift)
	off := abs & d.blockMask
	old := d.m[blk][off]
	d.m[blk][off] = v
	return old, true
}

// ReserveFront is an alias for Reserve(items, 0).
// See [Reserve].
func (d *Deque[T]) ReserveFront(items int) {
	d.Reserve(items, 0)
}

// ReserveBack is an alias for Reserve(0, items).
// See [Reserve].
func (d *Deque[T]) ReserveBack(items int) {
	d.Reserve(0, items)
}

// Reserve grows the deque's internal map of blocks (slice)
// to avoid further costs associated with map growth, but does not
// preallocate the blocks themselves.
//
// If the requested number of items can fit into the current capacity,
// Reserve is a noop. Otherwise it allocates a new map large enough
// to hold the existing blocks plus the blocks required for the requested
// number of items at each side, then copies the slice headers of existing
// blocks.
//
// The allocated map has capacity roughly equal to
// Len/BlockSize + (frontItems+backItems)/BlockSize.
// This uses much less memory than [EnsureFront] and [EnsureBack], since
// those also allocate the blocks themselves. Reserve eliminates
// up to log2((frontItems+backItems)/BlockSize) future map growths,
// assuming the map is nearly full when Reserve is called.
//
// If either of the arguments is negative, the function returns immediately.
func (d *Deque[T]) Reserve(frontItems, backItems int) {
	if frontItems < 0 || backItems < 0 {
		return
	}
	if frontItems == 0 && backItems == 0 {
		return
	}

	frontBlocks := d.blocksForCapCeil(frontItems)
	backBlocks := d.blocksForCapCeil(backItems)

	d.reserveBlocks(frontBlocks, backBlocks)
	// don't forget to reload d.m / d.front / d.back if cached
}

// TODO: add Ensure(front, back)

// EnsureFront grows the deque to allow at least the specified number of items
// to be pushed to the front before the next allocation occurs. Note that this
// does not change the empty space at the back, so if you need to push n items
// to the front and m items to the back before the next allocation, you should
// call EnsureFront(n) and by EnsureBack(m) together.
//
// See [EnsureBack] for the same operation at the back of the deque.
//
// If you want to reserve the space in the deque without eagerly allocating
// the blocks themselves, use [ReserveFront] or [Reserve] instead.
func (d *Deque[T]) EnsureFront(items int) {
	if items < 1 {
		return
	}
	blockSize := d.blockSize
	front := d.front
	if blockSize*front.blk+front.off < items {
		d.growFrontBy(items)
		// Required: reload front to not use stale copy.
		front = d.front
	}
	remaining := items
	for i := front.blk - 1; remaining > 0 && i >= 0; i-- {
		if d.m[i] == nil {
			d.m[i] = d.a.NewBlock()
		}
		remaining -= blockSize
	}
}

// EnsureBack grows the deque to allow at least the specified number of items
// to be pushed to the back before the next allocation occurs. Note that this
// does not change the empty space at the front, so if you need to push n items
// to the front and m items to the back before the next allocation, you should
// call EnsureFront(n) and by EnsureBack(m) together.
//
// See [EnsureFront] for the same operation at the front of the deque.
//
// If you want to reserve the space in the deque without eagerly allocating
// the blocks themselves, use [ReserveBack] or [Reserve] instead.
func (d *Deque[T]) EnsureBack(items int) {
	if items < 1 {
		return
	}
	back := d.back
	blockSize := d.blockSize
	if blockSize*(len(d.m)-1-back.blk)+(blockSize-back.off-1) < items {
		d.growBackBy(items)
		// Defensive: reload back
		back = d.back
	}
	remaining := items
	for i := back.blk + 1; remaining > 0 && i < len(d.m); i++ {
		if d.m[i] == nil {
			d.m[i] = d.a.NewBlock()
		}
		remaining -= blockSize
	}
}

// Preallocates enough blocks for the specified number of items in the pool.
// Returns true on success, false otherwise (if pooling is disabled).
func (d *Deque[T]) PreallocPool(items int) bool {
	if !d.a.pooled {
		return false
	}
	n := blocksForCapCeil(d.blockSize, items)
	d.a.PreallocBlocks(n)
	return true
}

// ShrinkToFit shrinks the deque to the max of (used blocks, sun of initially
// configured front and back slack). If pooling is used, the memory is not
// reclaimed by the pool - the deque and the pool will not hold any pointers
// to the underlying memory beyond the amount required by the above formula.
func (d *Deque[T]) ShrinkToFit() {
	if d.len == 0 {
		d.ClearRelease()
		return
	}
	initCfg := d.initCfg

	firstUsed := d.front.blk
	lastUsed := d.back.blk
	if d.back.off == 0 {
		lastUsed--
	}
	usedBlocks := lastUsed - firstUsed + 1

	slackFront := initCfg.frontBlock
	slackBack := initCfg.totalBlocks - initCfg.backBlock - 1

	newTotal := usedBlocks + slackFront + slackBack
	newMap := make([][]T, newTotal)
	copy(newMap[slackFront:], d.m[firstUsed:lastUsed+1])

	d.m = newMap
	d.front.blk = slackFront
	d.back.blk = slackFront + usedBlocks - 1
}

// Clear resets the deque without releasing any memory. This is useful
// when you want to reuse the deque without redundant reallocations.
// If the deque stores values and not pointers, make sure you don't have
// any live pointers to the items within the deque (create copies if you need)
// because the memory those pointers point to is up for reuse after this method
// is called. If you have many such live pointers, use [ClearRelease] instead.
func (d *Deque[T]) Clear() {
	d.resetToInit()
}

// ClearRelease discards all the allocated blocks and resets the deque
// to the initial state.
func (d *Deque[T]) ClearRelease() {
	d.resetToInit()
	initCfg := d.initCfg
	totalBlocks := initCfg.totalBlocks
	if cap(d.m) < totalBlocks {
		d.init()
		return
	}
	// INVARIANT: initCfg.frontBlock < len(d.m)
	b := d.m[initCfg.frontBlock]
	if b == nil {
		b = d.a.NewBlock()
	}
	newM := make([][]T, totalBlocks)
	newM[initCfg.frontBlock] = b
	d.m = newM
	d.a.ReleaseAll()
}

// Iter creates a forward iterator over the deque.
// Operations such as [Next], [Skip], and some others advance the iterator
// by one or more items. All operations work on the remaining portion
// of the deque.
func (d *Deque[T]) Iter() *Iter[T] {
	return &Iter[T]{b: d.newIter(d.front)}
}

// RevIter creates a reverse iterator over the deque.
// Operations such as [Next], [Skip], and some others advance the iterator
// by one or more items. All operations work on the remaining portion
// of the deque.
func (d *Deque[T]) RevIter() *RevIter[T] {
	b := d.newIter(d.back)
	b.pos = b.d.len
	return &RevIter[T]{b: b}
}

// BidiIter creates a bidirectional iterator over the deque.
// Operations such as [Next], [Skip], and some others advance the iterator
// by one or more items. All operations work on the remaining portion
// of the deque.
func (d *Deque[T]) BidiIter() *BidiIter[T] {
	bidi := d.newIter(d.front)
	return &bidi
}

// Slice creates a [start, end) slice of the deque if possible.
// On success, the length of the resulting slice is min(Len(), end-start).
// Returns an empty slice if the range is invalid.
func (d *Deque[T]) Slice(start, end int) []T {
	if start >= d.len || start < 0 {
		return []T{}
	}
	n := end - start
	if n < 1 {
		return []T{}
	}
	if 2*start >= d.len {
		return d.BidiIter().ResetBack().SkipBack(d.len - start).TakeSlice(n)
	}
	return d.Iter().Skip(start).TakeSlice(n)
}

// SlicePtr creates a [start, end) slice of the deque if possible.
// On success, the length of the resulting slice is min(Len(), end-start).
// Returns an empty slice if the range is invalid.
func (d *Deque[T]) PtrSlice(start, end int) []*T {
	if start >= d.len || start < 0 {
		return []*T{}
	}
	n := end - start
	if n < 1 {
		return []*T{}
	}
	if 2*start >= d.len {
		return d.BidiIter().ResetBack().SkipBack(d.len - start).TakePtrSlice(n)
	}
	return d.Iter().Skip(start).TakePtrSlice(n)
}

// ToSlice creates a slice with all the items of the deque.
func (d *Deque[T]) ToSlice() []T {
	return d.Slice(0, d.len)
}

// ToPtrSlice creates a slice with all the items of the deque.
func (d *Deque[T]) ToPtrSlice() []*T {
	return d.PtrSlice(0, d.len)
}

// ToChan creates a read channel with all the items of the deque.
func (d *Deque[T]) ToChan(size int) <-chan T {
	return d.Iter().ToChan(size)
}

// ToPtrChan creates a read channel with all the items of the deque.
func (d *Deque[T]) ToPtrChan(size int) <-chan *T {
	return d.Iter().ToPtrChan(size)
}

// Introspector returns a live introspector for the deque.
// The API provided by introspector is strictly for introspection during
// debugging or testing. It may change without notice.
// The introspector does not use any synchronization.
func (d *Deque[T]) Introspector() DequeIntrospector[T] {
	return DequeIntrospector[T]{d: d}
}

// ----- Helpers -----

func (d *Deque[T]) newIter(start blockIndex) BidiIter[T] {
	posFromFront := (d.front.blk << d.blockShift) + d.front.off
	return BidiIter[T]{
		d:        d.dequeState,
		blockCfg: d.blockCfg,
		cur:      start,
		frontAbs: posFromFront,
	}
}

func (d *Deque[T]) blocksForCapCeil(capItems int) int {
	return (capItems + d.blockMask) >> d.blockShift
}

func blocksForCapCeil(blockSize, capItems int) int {
	// This is only for init, hot path uses method with shifts.
	return (capItems + blockSize - 1) / blockSize
}

// init resets the deque and initializes the map according to Deque.initCfg
func (d *Deque[T]) init() {
	d.resetToInit()
	d.m = d.createMap(d.initCfg.totalBlocks, d.front.blk, 1)
}

// resetToInit resets the deque structure but does not release memory.
func (d *Deque[T]) resetToInit() {
	if d.a.pooled {
		for i, b := range d.m {
			if b == nil {
				continue
			}
			if i == d.initCfg.frontBlock {
				continue // keep this block
			}
			d.a.ReclaimBlock(b)
		}
	}
	d.front.blk = d.initCfg.frontBlock
	d.front.off = 0
	d.back.blk = d.initCfg.backBlock
	d.back.off = 0
	d.len = 0
}

func (d *Deque[T]) growFrontBy(n int) {
	back := d.back
	m := d.m
	if d.a.pooled {
		if i := back.blk + 1; i < len(d.m) {
			// Blocks after back+1 are reclaimed on PopBack.
			backNextBlk := m[i]
			if backNextBlk != nil {
				d.a.ReclaimBlock(backNextBlk)
			}
		}
	}

	allocBlkBeforeFront := 0
	for i := d.front.blk - 1; i >= 0; i-- {
		if m[i] == nil {
			break
		}
		allocBlkBeforeFront++
	}

	extraBlocks := d.blocksForCapCeil(n)
	d.reserveBlocks(extraBlocks, 0)
	// don't forget to reload d.m / d.front / d.back if cached
	m = d.m
	need := extraBlocks - allocBlkBeforeFront
	start := d.front.blk - allocBlkBeforeFront - 1
	for i := range need {
		m[start-i] = d.a.NewBlock()
	}
}

func (d *Deque[T]) growBackBy(n int) {
	front := d.front
	back := d.back
	m := d.m
	if d.a.pooled && front.blk > 0 {
		// Blocks before front+1 are reclaimed on PopFront.
		i := front.blk - 1
		frontPrevBlk := m[i]
		if frontPrevBlk != nil {
			d.a.ReclaimBlock(frontPrevBlk)
		}
	}

	allocBlkAfterBack := 0
	for i := back.blk + 1; i < len(d.m); i++ {
		if m[i] == nil {
			break
		}
		allocBlkAfterBack++
	}

	extraBlocks := d.blocksForCapCeil(n)
	d.reserveBlocks(0, extraBlocks)
	// don't forget to reload d.m / d.front / d.back if cached
	m = d.m

	need := extraBlocks - allocBlkAfterBack
	start := d.back.blk + allocBlkAfterBack + 1
	for i := range need {
		m[start+i] = d.a.NewBlock()
	}
}

// ReserveBlocks is the same as Reserve but in blocks and unchecked.
// Don't forget to reload d.m, d.front, and d.back if cached.
func (d *Deque[T]) reserveBlocks(frontBlocks, backBlocks int) {
	usedBlocks := d.back.blk - d.front.blk + 1
	frontSlack := d.front.blk
	backSlack := cap(d.m) - d.back.blk - 1
	if frontBlocks <= frontSlack && backBlocks <= backSlack {
		return // Items can fit without growing the map.
	}

	neededFront := numutil.MaxInt(frontSlack, frontBlocks)
	neededBack := numutil.MaxInt(backSlack, backBlocks)
	neededBlocks := neededFront + usedBlocks + neededBack
	newCap := numutil.MaxInt(cap(d.m)*2, neededBlocks)

	start := neededFront
	if frontBlocks > backBlocks {
		start = newCap - usedBlocks - neededBack
	}
	d.m = d.makeMapWithSlack(newCap, start, usedBlocks)

	d.front.blk = start
	d.back.blk = start + usedBlocks - 1
}

func (d *Deque[T]) makeMapWithSlack(newCap, start, usedBlocks int) [][]T {
	m := make([][]T, newCap)
	for i := range usedBlocks {
		m[start+i] = d.m[d.front.blk+i]
	}
	return m
}

// create new map and allocate nInitBlocks starting with firstInitBlock
func (d *Deque[T]) createMap(newCapBlocks, firstInitBlock, nInitBlocks int) [][]T {
	m := make([][]T, newCapBlocks)
	for i := firstInitBlock; i < firstInitBlock+nInitBlocks; i++ {
		m[i] = d.a.NewBlock()
	}
	return m
}

// reclaimBlock returns one block to the pool.
// Should not be called if not pooled.
func (d *Deque[T]) reclaimBlock(idx int) {
	block := d.m[idx]
	if block != nil {
		d.m[idx] = nil
		d.a.ReclaimBlock(block)
	}
}

// ----- Iter -----

// Clone creates a new forward iterator to the same deque,
// starting at the current position of this iterator.
func (it *Iter[T]) Clone() *Iter[T] {
	return &Iter[T]{b: *it.b.Clone()}
}

func (it *Iter[T]) Bidi() *BidiIter[T] {
	return &it.b
}

func (it *Iter[T]) Rev() *RevIter[T] {
	return &RevIter[T]{b: it.b}
}

// Reset sets this iterator to point to the front of the deque
// and returns the iterator itself.
func (it *Iter[T]) Reset() *Iter[T] {
	it.b.Reset()
	return it
}

// HasNext reports whether there is a next item.
//
// You only need to call this method explicitly if you want to check for
// remaining items yourself. Most iterator methods call it internally.
func (it *Iter[T]) HasNext() bool {
	return it.b.HasNext()
}

// Next returns the next item and true if the item is available,
// zero value and false otherwise. The iterator advances by one position.
func (it *Iter[T]) Next() (T, bool) {
	return it.b.Next()
}

// NextPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise. The iterator advances by one position.
func (it *Iter[T]) NextPtr() (*T, bool) {
	return it.b.NextPtr()
}

// Peek returns the next item and true if it is available,
// zero value and false otherwise.
func (it *Iter[T]) Peek() (T, bool) {
	return it.b.Peek()
}

// PeekPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise.
func (it *Iter[T]) PeekPtr() (*T, bool) {
	return it.b.PeekPtr()
}

// ForEach calls the specified function for each remaining item.
func (it *Iter[T]) ForEach(f func(T) bool) {
	it.b.ForEach(f)
}

// ForEachPtr calls the specified function for each remaining item.
func (it *Iter[T]) ForEachPtr(f func(val *T) bool) {
	it.b.ForEachPtr(f)
}

// ToChan consumes the iterator and returns a channel of specified size
// to which all next items starting from the current position will be written.
func (it *Iter[T]) ToChan(size int) <-chan T {
	return it.b.ToChan(size)
}

// ToPtrChan consumes the iterator and returns a read channel of specified size
// to which all next items starting from the current position will be written.
func (it *Iter[T]) ToPtrChan(size int) <-chan *T {
	return it.b.ToPtrChan(size)
}

// TakeSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *Iter[T]) TakeSlice(n int) []T {
	return it.b.TakeSlice(n)
}

// TakePtrSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *Iter[T]) TakePtrSlice(n int) []*T {
	return it.b.TakePtrSlice(n)
}

// TakeWhile collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns empty slice if the deque
// is empty.
func (it *Iter[T]) TakeWhile(pred func(T) bool) []T {
	return it.b.TakeWhile(pred)
}

// TakeWhilePtr collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns nil slice as empty.
// eq is a function that checks if two items are equal.
func (it *Iter[T]) TakeWhilePtr(pred func(*T) bool) []*T {
	return it.b.TakeWhilePtr(pred)
}

// Skip advances the iterator by a specified number of positions (or fewer
// if fewer items remain) and returns the iterator itself.
func (it *Iter[T]) Skip(n int) *Iter[T] {
	it.b.Skip(n)
	return it
}

// SkipWhile advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *Iter[T]) SkipWhile(pred func(T) bool) *Iter[T] {
	it.b.SkipWhile(pred)
	return it
}

// SkipWhilePtr advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *Iter[T]) SkipWhilePtr(pred func(*T) bool) *Iter[T] {
	it.b.SkipWhilePtr(pred)
	return it
}

// Seq creates a lazy sequence from this iterator.
func (it *Iter[T]) Seq() seq.Seq[T] {
	return seq.New(it.Next)
}

// PtrSeq creates a lazy sequence from this iterator.
func (it *Iter[T]) PtrSeq() seq.Seq[*T] {
	return seq.New(it.NextPtr)
}

// ----- RevIter -----

// Clone creates a new reverse iterator to the same deque,
// starting at the current position of this iterator.
func (it *RevIter[T]) Clone() *RevIter[T] {
	return &RevIter[T]{b: *it.b.Clone()}
}

func (it *RevIter[T]) Bidi() *BidiIter[T] {
	return &it.b
}

func (it *RevIter[T]) Rev() *Iter[T] {
	return &Iter[T]{b: it.b}
}

// Reset sets this iterator to point to the back of the deque
// and returns the iterator itself.
func (it *RevIter[T]) Reset() *RevIter[T] {
	it.b.ResetBack()
	return it
}

// HasNext reports whether there is a next item.
//
// You only need to call this method explicitly if you want to check for
// remaining items yourself. Most iterator methods call it internally.
func (it *RevIter[T]) HasNext() bool {
	return it.b.HasPrev()
}

// Next returns the next item and true if the item is available,
// zero value and false otherwise. The iterator advances by one position.
func (it *RevIter[T]) Next() (T, bool) {
	return it.b.Prev()
}

// NextPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise. The iterator advances by one position.
func (it *RevIter[T]) NextPtr() (*T, bool) {
	return it.b.PrevPtr()
}

// Peek returns the next item and true if it is available,
// zero value and false otherwise.
func (it *RevIter[T]) Peek() (T, bool) {
	return it.b.PeekBack()
}

// PeekPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise.
func (it *RevIter[T]) PeekPtr() (*T, bool) {
	return it.b.PeekBackPtr()
}

// ForEach calls the specified function for each remaining item.
func (it *RevIter[T]) ForEach(f func(T) bool) {
	for {
		v, ok := it.Next()
		if !ok || !f(v) {
			break
		}
	}
}

// ForEachPtr calls the specified function for each remaining item.
func (it *RevIter[T]) ForEachPtr(f func(val *T) bool) {
	for {
		v, ok := it.NextPtr()
		if !ok || !f(v) {
			break
		}
	}
}

// ToChan consumes the iterator and returns a channel of specified size
// to which all next items starting from the current position will be written.
func (it *RevIter[T]) ToChan(size int) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		for v, ok := it.Next(); ok; v, ok = it.Next() {
			ch <- v
		}
	}()
	return ch
}

// ToPtrChan consumes the iterator and returns a read channel of specified size
// to which all next items starting from the current position will be written.
func (it *RevIter[T]) ToPtrChan(size int) <-chan *T {
	ch := make(chan *T, size)
	go func() {
		defer close(ch)
		for v, ok := it.NextPtr(); ok; v, ok = it.NextPtr() {
			ch <- v
		}
	}()
	return ch
}

// TakeSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *RevIter[T]) TakeSlice(n int) []T {
	if !it.HasNext() || n < 1 {
		return []T{}
	}
	res := make([]T, 0, numutil.MinInt(it.b.d.len, n))
	i := 0
	for v, ok := it.Peek(); ok && i < n; v, ok = it.Peek() {
		res = append(res, v)
		_, _ = it.Next()
		i++
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakePtrSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *RevIter[T]) TakePtrSlice(n int) []*T {
	if !it.HasNext() || n < 1 {
		return []*T{}
	}
	res := make([]*T, 0, numutil.MinInt(it.b.d.len, n))
	i := 0
	for v, ok := it.PeekPtr(); ok && i < n; v, ok = it.PeekPtr() {
		res = append(res, v)
		_, _ = it.Next()
		i++
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// TakeWhile collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns empty slice if the deque
// is empty.
func (it *RevIter[T]) TakeWhile(pred func(T) bool) []T {
	if !it.HasNext() {
		return []T{}
	}
	res := make([]T, 0, numutil.MinInt(it.b.d.len, takeWhileInitCap))
	for v, ok := it.Peek(); ok && pred(v); v, ok = it.Peek() {
		res = append(res, v)
		_, _ = it.Next()
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakeWhilePtr collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns nil slice as empty.
// eq is a function that checks if two items are equal.
func (it *RevIter[T]) TakeWhilePtr(pred func(*T) bool) []*T {
	if !it.HasNext() {
		return []*T{}
	}
	res := make([]*T, 0, numutil.MinInt(it.b.d.len, takeWhileInitCap))
	for v, ok := it.PeekPtr(); ok && pred(v); v, ok = it.PeekPtr() {
		res = append(res, v)
		_, _ = it.Next()
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// Skip advances the iterator by a specified number of positions (or fewer
// if fewer items remain) and returns the iterator itself.
func (it *RevIter[T]) Skip(n int) *RevIter[T] {
	it.b.SkipBack(n)
	return it
}

// SkipWhile advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *RevIter[T]) SkipWhile(pred func(T) bool) *RevIter[T] {
	for v, _ := it.Peek(); it.HasNext() && pred(v); v, _ = it.Peek() {
		it.b.stepBack()
	}
	return it
}

// SkipWhilePtr advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *RevIter[T]) SkipWhilePtr(pred func(*T) bool) *RevIter[T] {
	for v, _ := it.PeekPtr(); it.HasNext() && pred(v); v, _ = it.PeekPtr() {
		it.b.stepBack()
	}
	return it
}

// Seq creates a lazy sequence from this iterator.
func (it *RevIter[T]) Seq() seq.Seq[T] {
	return seq.New(it.Next)
}

// PtrSeq creates a lazy sequence from this iterator.
func (it *RevIter[T]) PtrSeq() seq.Seq[*T] {
	return seq.New(it.NextPtr)
}

// ----- BidiIter -----

// Clone creates a new bidirectional iterator to the same deque,
// starting at the current position of this iterator.
func (it *BidiIter[T]) Clone() *BidiIter[T] {
	clone := *it
	return &clone
}

// Reset sets this iterator to point to the front of the deque
// and returns the iterator itself.
func (it *BidiIter[T]) Reset() *BidiIter[T] {
	it.cur = it.d.front
	it.pos = 0
	return it
}

// ResetBack sets this iterator to point to the back of the deque
// and returns the iterator itself.
func (it *BidiIter[T]) ResetBack() *BidiIter[T] {
	it.cur = it.d.back
	it.pos = it.d.len
	return it
}

// HasNext reports whether there is a next item.
//
// You only need to call this method explicitly if you want to check for
// remaining items yourself. Most iterator methods call it internally.
func (it *BidiIter[T]) HasNext() bool {
	return it.pos < it.d.len
}

// HasPrev reports whether there is a previous item.
func (it *BidiIter[T]) HasPrev() bool {
	return it.pos > 0
}

// Next returns the next item and true if the item is available,
// zero value and false otherwise. The iterator advances by one position.
func (it *BidiIter[T]) Next() (T, bool) {
	v, ok := it.Peek()
	if ok {
		it.advance()
	}
	return v, ok
}

// NextPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise. The iterator advances by one position.
func (it *BidiIter[T]) NextPtr() (*T, bool) {
	v, ok := it.PeekPtr()
	if ok {
		it.advance()
	}
	return v, ok
}

// Prev returns the previous item and true if the item is available,
// zero value and false otherwise.
func (it *BidiIter[T]) Prev() (T, bool) {
	v, ok := it.PeekBack()
	if ok {
		it.stepBack()
	}
	return v, ok
}

// PrevPtr returns the previous item and true if the item is available,
// zero value and false otherwise.
func (it *BidiIter[T]) PrevPtr() (*T, bool) {
	v, ok := it.PeekBackPtr()
	if ok {
		it.stepBack()
	}
	return v, ok
}

// Peek returns the next item and true if it is available,
// zero value and false otherwise.
func (it *BidiIter[T]) Peek() (T, bool) {
	if !it.HasNext() {
		var zero T
		return zero, false
	}
	return it.d.m[it.cur.blk][it.cur.off], true
}

// PeekPtr returns a pointer to the next item and true if it is available,
// nil and false otherwise.
func (it *BidiIter[T]) PeekPtr() (*T, bool) {
	if !it.HasNext() {
		return nil, false
	}
	return &it.d.m[it.cur.blk][it.cur.off], true
}

// PeekBack returns the previous item and true if it is available,
// zero value and false otherwise.
func (it *BidiIter[T]) PeekBack() (T, bool) {
	if !it.HasPrev() {
		var zero T
		return zero, false
	}
	blk := it.cur.blk
	off := it.cur.off - 1
	if off < 0 {
		blk--
		off = it.blockSize - 1
	}
	return it.d.m[blk][off], true
}

// PeekBackPtr returns a pointer to the previous item and true
// if it is available, nil and false otherwise.
func (it *BidiIter[T]) PeekBackPtr() (*T, bool) {
	if !it.HasPrev() {
		return nil, false
	}
	blk := it.cur.blk
	off := it.cur.off - 1
	if off < 0 {
		blk--
		off = it.blockSize - 1
	}
	return &it.d.m[blk][off], true
}

// ForEach calls the specified function for each remaining item.
func (it *BidiIter[T]) ForEach(f func(T) bool) {
	for {
		v, ok := it.Next()
		if !ok || !f(v) {
			break
		}
	}
}

// ForEachPtr calls the specified function for each remaining item.
func (it *BidiIter[T]) ForEachPtr(f func(val *T) bool) {
	for {
		v, ok := it.NextPtr()
		if !ok || !f(v) {
			break
		}
	}
}

// ToChan consumes the iterator and returns a channel of specified size
// to which all next items starting from the current position will be written.
func (it *BidiIter[T]) ToChan(size int) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		for v, ok := it.Next(); ok; v, ok = it.Next() {
			ch <- v
		}
	}()
	return ch
}

// ToPtrChan consumes the iterator and returns a read channel of specified size
// to which all next items starting from the current position will be written.
func (it *BidiIter[T]) ToPtrChan(size int) <-chan *T {
	ch := make(chan *T, size)
	go func() {
		defer close(ch)
		for v, ok := it.NextPtr(); ok; v, ok = it.NextPtr() {
			ch <- v
		}
	}()
	return ch
}

// TakeSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *BidiIter[T]) TakeSlice(n int) []T {
	if n < 1 || !it.HasNext() {
		return []T{}
	}
	res := make([]T, 0, numutil.MinInt(it.d.len, n))
	for range n {
		v, ok := it.Next()
		if !ok {
			break
		}
		res = append(res, v)
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakePtrSlice collects up to n items to a slice, starting from the next
// item. Returns an empty slice if the deque is empty.
func (it *BidiIter[T]) TakePtrSlice(n int) []*T {
	if n < 1 || !it.HasNext() {
		return []*T{}
	}
	res := make([]*T, 0, numutil.MinInt(it.d.len, n))
	for range n {
		v, ok := it.NextPtr()
		if !ok {
			break
		}
		res = append(res, v)
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// TakeWhile collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns empty slice if the deque
// is empty.
func (it *BidiIter[T]) TakeWhile(pred func(T) bool) []T {
	if !it.HasNext() {
		return []T{}
	}
	res := make([]T, 0, numutil.MinInt(it.d.len, takeWhileInitCap))
	for v, ok := it.Peek(); ok && pred(v); v, ok = it.Peek() {
		res = append(res, v)
		it.advance()
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []T{}
	}
	return res
}

// TakeWhilePtr collects items to a slice until it encounters the first item that
// does not match the specified predicate. Returns nil slice as empty.
// eq is a function that checks if two items are equal.
func (it *BidiIter[T]) TakeWhilePtr(pred func(*T) bool) []*T {
	if !it.HasNext() {
		return []*T{}
	}
	res := make([]*T, 0, numutil.MinInt(it.d.len, takeWhileInitCap))
	for v, ok := it.PeekPtr(); ok && pred(v); v, ok = it.PeekPtr() {
		res = append(res, v)
		it.advance()
	}
	if len(res) == 0 { // defensive check, should never be true.
		return []*T{}
	}
	return res
}

// Skip advances the iterator by a specified number of positions (or fewer
// if fewer items remain) and returns the iterator itself.
func (it *BidiIter[T]) Skip(n int) *BidiIter[T] {
	if n < 1 || !it.HasNext() {
		return it
	}
	if n > it.d.len-it.pos {
		return it.ResetBack()
	}
	it.moveUnsafe(n)
	return it
}

// SkipBack steps back the iterator by a specified number of positions (or fewer
// if fewer items are before the current position) and returns the iterator
// itself.
func (it *BidiIter[T]) SkipBack(n int) *BidiIter[T] {
	if n < 1 || !it.HasPrev() {
		return it
	}
	if n > it.pos {
		return it.Reset()
	}
	it.moveUnsafe(-n)
	return it
}

// SkipWhile advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *BidiIter[T]) SkipWhile(pred func(T) bool) *BidiIter[T] {
	for v, _ := it.Peek(); it.HasNext() && pred(v); v, _ = it.Peek() {
		it.advance()
	}
	return it
}

// SkipWhilePtr advances the iterator until it encounters the first item that
// does not match the specified predicate. Returns the iterator itself
// positioned just before the first item that fails the predicate,
// which will be returned by the first call to [Next].
func (it *BidiIter[T]) SkipWhilePtr(pred func(*T) bool) *BidiIter[T] {
	for v, _ := it.PeekPtr(); it.HasNext() && pred(v); v, _ = it.PeekPtr() {
		it.advance()
	}
	return it
}

// Seq creates a lazy sequence from this iterator.
func (it *BidiIter[T]) Seq() seq.Seq[T] {
	return seq.New(it.Next)
}

// PtrSeq creates a lazy sequence from this iterator.
func (it *BidiIter[T]) PtrSeq() seq.Seq[*T] {
	return seq.New(it.NextPtr)
}

// RevSeq creates a reverse lazy sequence from this iterator.
func (it *BidiIter[T]) RevSeq() seq.Seq[T] {
	rev := it.Clone().ResetBack()
	return seq.New(rev.Prev)
}

// RevPtrSeq creates a reverse lazy sequence from this iterator.
func (it *BidiIter[T]) RevPtrSeq() seq.Seq[*T] {
	rev := it.Clone().ResetBack()
	return seq.New(rev.PrevPtr)
}

// moveUnsafe moves the current iterator n positions to the direction dictated
// by the sign of n.
func (it *BidiIter[T]) moveUnsafe(n int) {
	it.pos += n
	abs := it.frontAbs + it.pos
	it.cur.blk = abs >> it.blockShift
	it.cur.off = abs & it.blockMask
}

func (it *BidiIter[T]) advance() {
	it.cur.off++
	if it.cur.off >= it.blockSize {
		it.cur.blk++
		it.cur.off = 0
	}
	it.pos++
}

func (it *BidiIter[T]) stepBack() {
	it.cur.off--
	if it.cur.off < 0 {
		it.cur.blk--
		it.cur.off = it.blockSize - 1
	}
	it.pos--
}
