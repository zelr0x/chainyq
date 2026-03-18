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

type Sequencer[T any] interface {
	Seq() seq.Seq[T]
	PtrSeq() seq.Seq[*T]
}

type BidiSequencer[T any] interface {
	Sequencer[T]
	RevSeq() seq.Seq[T]
	RevpTrSeq() seq.Seq[*T]
}

type SeqCollection[T any] interface {
	Collection[T]
	Sequencer[T]
	ToSlice() []T
	ToPtrSlice() []*T
}

type BidiCollection[T any] interface {
	SeqCollection[T]
	BidiSequencer[T]
	RevSeq() seq.Seq[T]
	RevPtrSeq() seq.Seq[*T]
}

type RandomAccess[T any] interface {
	Collection[T]
	Get(int) (T, bool)
	Set(int, T) (T, bool)
}

type Queue[T any] interface {
	SeqCollection[T]
	PushBack(T)
	PopFront() (T, bool)
	Front() (T, bool)
	FrontPtr() (*T, bool)
}

type Stack[T any] interface {
	SeqCollection[T]
	PushBack(T)
	PopBack() (T, bool)
	Back() (T, bool)
	BackPtr() (*T, bool)
}

type Deque[T any] interface {
	BidiCollection[T]
	RandomAccess[T]

	PushBack(T)
	PopBack() (T, bool)
	Back() (T, bool)
	BackPtr() (*T, bool)

	PushFront(T)
	PopFront() (T, bool)
	Front() (T, bool)
	FrontPtr() (*T, bool)
}

type Nexter[T any] interface {
	Next() (T, bool)
}

type Peeker[T any] interface {
	Peek() (T, bool)
	PeekPtr() (*T, bool)
	HasNext() bool
}

type Iterator[T any] interface {
	Sequencer[T]
	Peeker[T]
	Nexter[T]
	NextPtr() (*T, bool)
}

type BidiIterator[T any] interface {
	BidiSequencer[T]
	Iterator[T]
	Prev() (T, bool)
	PrevPtr() (*T, bool)
}

type CursorIterator[T any] interface {
	Iterator[T]
	Current() (T, bool)
	CurrentPtr() (*T, bool)
}
