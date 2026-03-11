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

// Package chainyq provides generic data structures and iterator utilities.
//
// The core types are concrete and optimized for speed. Interfaces
// such as [Deque] or [Iterator] are defined for convenience.
// Note that using interfaces introduces dynamic dispatch, which can be
// noticeably slower in some cases, even when used as generic type parameter
// constraint. Prefer concrete types for performance-sensitive cases.
package chainyq

import "github.com/zelr0x/chainyq/seq"

type Collection[T any] interface {
	Len() int
	IsEmpty() bool
}

type Deque[T any] interface {
	Collection[T]

	PushBack(T)
	PopBack() (T, bool)
	Back() (T, bool)

	PushFront(T)
	PopFront() (T, bool)
	Front() (T, bool)
}

type Nexter[T any] interface {
	Next() (T, bool)
}

type Peeker[T any] interface {
	Peek() (T, bool)
	HasNext() bool
}

type Iterator[T any] interface {
	Nexter[T]
	NextPtr() (*T, bool)
	Peeker[T]
}

type CursorIterator[T any] interface {
	Iterator[T]
	Current() (T, bool)
	CurrentPtr() (*T, bool)
}

type Sequencer[T any] interface {
	Seq() seq.Seq[T]
	PtrSeq() seq.Seq[*T]
}
