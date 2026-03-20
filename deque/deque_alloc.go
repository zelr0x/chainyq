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

import (
	"github.com/zelr0x/chainyq/stack"
)

// separating dequeAlloc and blockAlloc allows swapping impls
// or hybrid strategies with multiple allocators.
type dequeAlloc[T any] struct {
	pool      stack.Stack[[]T]
	blockSize int
	pooled    bool
}

type dequeAllocCfg struct {
	blockSize int
	pooled    bool
}

func newDequeAlloc[T any](cfg dequeAllocCfg) dequeAlloc[T] {
	blockSize := cfg.blockSize
	a := dequeAlloc[T]{
		blockSize: blockSize,
		pooled:    cfg.pooled,
	}
	if cfg.pooled {
		a.pool = stack.NewValue[[]T](16)
	}
	return a
}

// Should not be called if not pooled.
func (a *dequeAlloc[T]) PreallocBlocks(n int) {
	a.pool.Ensure(n)
	blockSize := a.blockSize
	for range n {
		a.pool.Push(Allocate[T](blockSize))
	}
}

func (a *dequeAlloc[T]) NewBlock() []T {
	if !a.pooled {
		return Allocate[T](a.blockSize)
	}
	v, ok := a.pool.Pop()
	if !ok {
		return Allocate[T](a.blockSize)
	}
	return v
}

// Should not be called if not pooled.
func (a *dequeAlloc[T]) ReclaimBlock(block []T) {
	a.pool.Push(block)
}

func (a *dequeAlloc[T]) ReleaseAll() {
	if a.pooled {
		a.pool.Clear()
	}
}

func Allocate[T any](blockSize int) []T {
	return make([]T, blockSize)
}
