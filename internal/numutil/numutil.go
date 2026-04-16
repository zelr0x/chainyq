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

package numutil

import (
	"math"
	"math/bits"
)

func ClampInt(x, minVal, maxVal int) int {
	if x < minVal {
		return minVal
	}
	if x > maxVal {
		return maxVal
	}
	return x
}

func RoundNextPow2(x uint) uint {
	if x == 0 {
		return 1
	}
	return 1 << (bits.Len(x - 1))
}

func U64ToInt(u uint64) int {
	if u > uint64(math.MaxInt) {
		return math.MaxInt
	}
	return int(u)
}
