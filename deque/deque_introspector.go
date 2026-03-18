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

package deque

// DequeIntrospector provides live access to a deque's internal state.
// This API is strictly for introspection during debugging or testing.
// and it may change without notice.
type DequeIntrospector[T any] struct {
	d *Deque[T]
}

func (di DequeIntrospector[T]) Deque() *Deque[T] {
	return di.d
}

// Len returns the number of elements in the deque.
func (di DequeIntrospector[T]) Len() int {
	return di.d.len
}

// FrontIndex returns the front block/off index.
func (di DequeIntrospector[T]) FrontIndex() (blk, off int) {
	return di.d.front.blk, di.d.front.off
}

// BackIndex returns the back block/off index.
func (di DequeIntrospector[T]) BackIndex() (blk, off int) {
	return di.d.back.blk, di.d.back.off
}

// BlockSize returns the deque block size.
func (di DequeIntrospector[T]) BlockSize() int {
	return di.d.blockSize
}

// BlockMask returns the deque block mask.
func (di DequeIntrospector[T]) BlockMask() int {
	return di.d.blockMask
}

// BlockShift returns the deque block shift.
func (di DequeIntrospector[T]) BlockShift() uint {
	return di.d.blockShift
}

// Map returns a copy of the underlying block slice.
func (di DequeIntrospector[T]) Map() [][]T {
	return di.d.m
}

// Alloc returns whether the deque is pooled and its block size.
// You can expose more alloc info here if needed.
func (di DequeIntrospector[T]) AllocInfo() (pooled bool, blockSize int) {
	return di.d.a.pooled, di.d.a.blockSize
}

// InitConfig exposes scalar fields from initCfg
func (di DequeIntrospector[T]) InitConfig() (totalBlocks, frontBlock, backBlock int) {
	return di.d.initCfg.totalBlocks, di.d.initCfg.frontBlock, di.d.initCfg.backBlock
}
