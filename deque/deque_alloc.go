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

// separating dequeAlloc and blockAlloc allows swapping impls
// or hybrid strategies with multiple allocators.
type dequeAlloc[T any] struct {
	ba        blockAlloc[T]
	blockSize int
	pooled    bool
}

type dequeAllocCfg struct {
	blockSize     int
	initCap       int
	preallocItems int
	pooled        bool
}

func newDequeAlloc[T any](cfg dequeAllocCfg) dequeAlloc[T] {
	preallocBlocks := 0
	if cfg.preallocItems > 0 {
		preallocBlocks = blocksForCapCeil(cfg.blockSize, cfg.preallocItems)
	}
	ba := newBlockAlloc[T](blockAllocCfg{
		blockSize:      cfg.blockSize,
		initCap:        cfg.initCap,
		preallocBlocks: preallocBlocks,
	})
	return dequeAlloc[T]{
		pooled:    cfg.pooled,
		blockSize: cfg.blockSize,
		ba:        ba,
	}
}

func (a *dequeAlloc[T]) NewBlock() []T {
	if a.pooled {
		block, ok := a.ba.NewBlock()
		if ok {
			return block
		}
	}
	return a.allocate()
}

// It is a separate function to distinguish it easier during profiling.
// The call should be inlined in all cases, so not a big deal.
func (a *dequeAlloc[T]) allocate() []T {
	return make([]T, a.blockSize)
}

// Should not be called if not pooled.
func (a *dequeAlloc[T]) ReclaimBlock(block []T) {
	if !a.pooled {
		return
	}
	a.ba.Reclaim(block)
}

// Should not be called if not pooled.
func (a *dequeAlloc[T]) ReclaimChunk(chunk [][]T) {
	if !a.pooled {
		return
	}
	for _, v := range chunk {
		if v != nil {
			a.ba.Reclaim(v)
		}
	}
}

func (a *dequeAlloc[T]) ReleaseAll() {
	if !a.pooled {
		return
	}
	a.ba.ReleaseAll()
}
