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

package syncdeque

import "testing"

func BenchmarkConcurrentPushPop(b *testing.B) {
	d := New[int]()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				d.PushBack(i)
			case 1:
				d.PushFront(i)
			case 2:
				d.PopBack()
			case 3:
				d.PopFront()
			}
			i++
		}
	})
}

func BenchmarkReaderHeavy(b *testing.B) {
	d := New[int]()
	for i := range 1000 {
		d.PushBack(i)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = d.Len()
			_, _ = d.Front()
			_, _ = d.Back()
		}
	})
}

func BenchmarkWriterHeavy(b *testing.B) {
	d := New[int]()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			d.PushBack(i)
			d.PopFront()
			i++
		}
	})
}

func BenchmarkBurstyWithEnsure(b *testing.B) {
	d := New[int]()

	const burstSize = 100
	const opsPerIter = 2 * burstSize
	d.EnsureBack(burstSize)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			for range burstSize {
				d.PushBack(i)
				i++
			}
			for range burstSize {
				d.PopFront()
			}
		}
	})

	elapsed := b.Elapsed().Nanoseconds()
	nsPerDqOp := float64(elapsed) / float64(b.N*opsPerIter)
	b.ReportMetric(nsPerDqOp, "ns/deque-op")
}
