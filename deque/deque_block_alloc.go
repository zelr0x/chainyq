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

import "github.com/zelr0x/chainyq/stack"

type blockAllocCfg struct {
	blockSize      int
	initCap        int
	preallocBlocks int
}

type blockAlloc[T any] struct {
	s stack.Stack[[]T]
}

func newBlockAlloc[T any](cfg blockAllocCfg) blockAlloc[T] {
	stack := stack.NewValue[[]T](cfg.initCap)
	if cfg.preallocBlocks > 0 {
		preallocBlocks(&stack, cfg.preallocBlocks, cfg.blockSize)
	}
	return blockAlloc[T]{
		s: stack,
	}
}

func (a *blockAlloc[T]) NewBlock() ([]T, bool) {
	return a.s.Pop()
}

// Reclaim pools the block if pooled, otherwise it's a noop.
// The specified block should not be nil.
// Should not be called if not pooled.
func (a *blockAlloc[T]) Reclaim(block []T) {
	a.s.Push(block)
}

func (a *blockAlloc[T]) ReleaseAll() {
	a.s.Clear()
}

func preallocBlocks[T any](s *stack.Stack[[]T], blocks, blockSize int) {
	for range blocks {
		s.Push(make([]T, blockSize))
	}
}
