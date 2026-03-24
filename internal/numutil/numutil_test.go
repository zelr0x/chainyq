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
	"math/bits"
	"testing"

	. "github.com/zelr0x/chainyq/internal/testutil"
)

func TestClampInt(t *testing.T) {
	tests := []struct {
		name        string
		x, min, max int
		want        int
	}{
		{"below min", -5, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"within range", 5, 0, 10, 5},
		{"equal to min", 0, 0, 10, 0},
		{"equal to max", 10, 0, 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampInt(tt.x, tt.min, tt.max)
			AssertEq(t, tt.want, got)
		})
	}
}

func TestRoundNextPow2(t *testing.T) {
	AssertEq(t, 1, RoundNextPow2(0), "must be 1 for 0")

	for i := range bits.UintSize {
		x := uint(1) << i

		got2 := RoundNextPow2(x)
		AssertEq(t, x, got2, "must be x for x")

		if x >= 4 {
			got1 := RoundNextPow2(x - 1)
			AssertEq(t, x, got1, "must be x for all x-1 where x >= 4")

			if i+1 < bits.UintSize {
				got3 := RoundNextPow2(x>>1 + 1)
				AssertEq(t, x, got3, "must be x for the prevPow2+1 where x >= 4")
			}
		}
	}

	// This is a redundant defensive check, because unsafe ops may depend on it.
	j := 2 // 2**2 == 4, start with 3 and go all the way to 4096, including
	for i := uint(3); i <= uint(4096); i++ {
		got := RoundNextPow2(i)
		want := uint(1 << j)
		AssertEq(t, want, got)
		if (i & (i - 1)) == 0 { // if i is power of 2, go to the next one
			j++
		}
	}
}
