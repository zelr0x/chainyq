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

package listdeque_test

import (
	"fmt"
	"testing"

	deque "github.com/zelr0x/chainyq/deque/listdeque"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestString(t *testing.T) {
	var d *deque.ListDeque[int]
	AssertEq(t, "nil", d.String())
	d = deque.New[int]()
	AssertEq(t, "ListDeque[]", d.String())
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)
	AssertEq(t, "ListDeque[1 2 3]", d.String())
}

func TestGoString(t *testing.T) {
	var d *deque.ListDeque[int]
	AssertEq(t, "nil", d.String())
	d = deque.New[int]()
	AssertEq(t, "ListDeque[int]{}", fmt.Sprintf("%#v", d))
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)
	AssertEq(t, "ListDeque[int]{1, 2, 3}", fmt.Sprintf("%#v", d))
}

func TestPushFrontAndPushBack(t *testing.T) {
	l := deque.New[int]()
	l.PushBack(2)
	l.PushFront(1)
	l.PushBack(3)
	front, _ := l.Front()
	back, _ := l.Back()
	AssertEq(t, 1, front)
	AssertEq(t, 3, back)
	AssertEq(t, 3, l.Len())
}
