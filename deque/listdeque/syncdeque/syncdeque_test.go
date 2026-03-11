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

package syncdeque_test

import (
	"testing"
	"time"

	deque "github.com/zelr0x/chainyq/deque/listdeque/syncdeque"
)

// go test -race -timeout 30s -run ^TestConcurrentAccess
func TestConcurrentAccess(t *testing.T) {
	d := deque.New[int]()
	const N = 10000

	// PopFront + PushBack
	g1 := func() {
		for i := 0; i < N; i++ {
			if v, ok := d.PopFront(); ok {
				d.PushBack(v)
			} else {
				d.PushBack(i)
			}
		}
	}

	// PopBack + PushFront
	g2 := func() {
		for i := 0; i < N; i++ {
			if v, ok := d.PopBack(); ok {
				d.PushFront(v)
			} else {
				d.PushFront(i)
			}
		}
	}

	// Reads
	g3 := func() {
		for i := 0; i < N; i++ {
			_ = d.Len()
			_ = d.IsEmpty()
			d.Front()
			d.Back()
		}
	}

	done := make(chan struct{})
	go func() { g1(); close(done) }()
	go g2()
	go g3()

	select {
	case <-done:
		// success — one goroutine finished, others still running
	case <-time.After(5 * time.Second):
		t.Fatal("possible deadlock or hang")
	}
}
