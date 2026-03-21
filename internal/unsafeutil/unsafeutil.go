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

package unsafeutil

import "unsafe"

// i must not be negative.
func At[T any](base *T, i int) *T {
	return (*T)(unsafe.Add(unsafe.Pointer(base), uintptr(i)*unsafe.Sizeof(*base))) // #nosec G103 G115
}
