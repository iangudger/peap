// Copyright 2018 Google LLC
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

package peap

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestLog2(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{1, 0},
		{2, 1},
		{3, 1},
		{4, 2},
		{5, 2},
		{6, 2},
		{7, 2},
		{8, 3},
		{9, 3},
	}
	for _, test := range tests {
		if got := log2(test.in); got != test.want {
			t.Errorf("got Log2(%d) = %d, want = %d", test.in, got, test.want)
		}
	}
}

type element struct {
	Entry
	value uint32
}

func (e *element) Less(elem Element) bool {
	return e.value < elem.(*element).value
}

func (e *element) String() string {
	return fmt.Sprintf("%03d", e.value)
}

func TestInsert(t *testing.T) {
	min := uint32(math.MaxUint32)
	rand.Seed(time.Now().Unix())

	var h Heap

	for i := 0; i < 100; i++ {
		cur := rand.Uint32() % 1000
		if cur < min {
			min = cur
		}
		h.Push(&element{value: cur})
	}

	if got := h.root.(*element).value; got != min {
		t.Errorf("got root = %d, want = %d", got, min)
	}
}

func TestRemove(t *testing.T) {
	rand.Seed(time.Now().Unix())

	values := make(sort.IntSlice, 0, 100)
	var h Heap
	for i := 0; i < 100; i++ {
		cur := rand.Uint32() % 1000
		values = append(values, int(cur))
		h.Push(&element{value: cur})
	}

	values.Sort()
	for len(values) > 0 {
		got := int(h.Pop().(*element).value)
		if got != values[0] {
			t.Errorf("got h.Pop() = %d, want = %d", got, values[0])
		}
		values = values[1:]
	}

	if got := h.Len(); got != 0 {
		t.Errorf("removed all elements, got h.Len() = %d, want = 0", got)
	}
	if h.root != nil {
		t.Errorf("removed all elements, got h.root = %v, want = nil", h.root)
	}
}
