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
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
	Entry[*element]
	value uint32
}

func (e *element) Less(elem *element) bool {
	return e.value < elem.value
}

func (e *element) String() string {
	return fmt.Sprintf("%05d", e.value)
}

func TestInsert(t *testing.T) {
	min := uint32(math.MaxUint32)

	var h Heap[*element]

	for i := 0; i < 100; i++ {
		cur := rand.Uint32() % 1000
		if cur < min {
			min = cur
		}
		h.Push(&element{value: cur})
	}

	if got := h.root.value; got != min {
		t.Errorf("got root = %d, want = %d", got, min)
	}
}

func TestRemove(t *testing.T) {
	values := make(sort.IntSlice, 0, 100)
	var h Heap[*element]
	for i := 0; i < 100; i++ {
		cur := rand.Uint32() % 1000
		values = append(values, int(cur))
		h.Push(&element{value: cur})
	}

	values.Sort()
	for len(values) > 0 {
		got := int(h.Pop().value)
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

func TestString(t *testing.T) {
	var h Heap[*element]

	for i := 0; i < 10; i++ {
		h.Push(&element{value: uint32(i)})
	}

	got := h.String()
	const want = `00000
00001 00002
00003 00004 00005 00006
00007 00008 00009 <nil> <nil> <nil> <nil> <nil>
<nil> <nil> <nil> <nil> <nil> <nil>
`

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("String() mismatch (-want +got):\n%s", diff)
	}
}

func BenchmarkHeap(b *testing.B) {
	for _, size := range []int{5, 10, 100} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				values := make([]*element, 0, size)
				for i := 0; i < cap(values); i++ {
					values = append(values, &element{value: rand.Uint32() % 1000})

				}

				var h Heap[*element]
				b.StartTimer()

				for _, v := range values {
					h.Push(v)
				}
				for h.Pop() != nil {
				}
			}
		})
	}
}

type uint32Heap []uint32

func (h uint32Heap) Len() int           { return len(h) }
func (h uint32Heap) Less(i, j int) bool { return h[i] < h[j] }
func (h uint32Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *uint32Heap) Push(x any) {
	*h = append(*h, x.(uint32))
}

func (h *uint32Heap) Pop() any {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[0 : n-1]
	return x
}

func BenchmarkContainerHeap(b *testing.B) {
	for _, size := range []int{5, 10, 100} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				values := make([]uint32, 0, size)
				for i := 0; i < cap(values); i++ {
					values = append(values, rand.Uint32()%1000)

				}

				var h uint32Heap
				heap.Init(&h)
				b.StartTimer()

				for _, v := range values {
					heap.Push(&h, v)
				}
				for h.Len() > 0 {
					heap.Pop(&h)
				}
			}
		})
	}
}

func init() {
	rand.Seed(time.Now().Unix())
}
