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

package benchutil

import (
	"math/rand"
	"testing"
)

func RandomIntSlice(b *testing.B, seed int64, intn int) []int {
	b.Helper()
	return RandomIntSliceN(b, seed, b.N, intn)
}

func RandomIntSliceN(b *testing.B, seed int64, size, intn int) []int {
	b.Helper()
	rand := rand.New(rand.NewSource(seed))
	a := make([]int, size)
	for i := range size {
		a[i] = rand.Intn(intn)
	}
	return a
}
