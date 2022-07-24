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

// Package peap provides the implementation of a pointer-based (array-less)
// intrusive min-heap. Peap is a portmanteau of pointer heap.
//
// Recommended usage: Attach a little bo.
//
// This package is based on the paper,"Peaps: Heaps implemented without
// arrays":
// https://www.cpp.edu/~ftang/courses/CS241/notes/Building_Heaps_With_Pointers.pdf
package peap

import (
	"fmt"
	"math/bits"
)

// Linker is the interface that objects must implement if they want to be added
// to and/or removed from Heap objects.
//
// N.B. When substituted in a template instantiation, Linker doesn't need to
// be an interface, and in most cases won't be.
type Linker[T any] interface {
	Left() T
	Right() T
	SetLeft(T)
	SetRight(T)
}

// Element the item that is used at the API level.
//
// N.B. Like Linker, this is unlikely to be an interface in most cases. If
// Element is not an interface, it must be a pointer.
type Element[T any] interface {
	comparable
	Linker[T]
	Less(T) bool
}

// Heap implement a pointer-based min-heap.
type Heap[T Element[T]] struct {
	size int
	root T
}

// Peek returns the next Element to be removed from the Heap.
func (h *Heap[T]) Peek() T {
	return h.root
}

// Len returns the number of Elements currently in the Heap.
func (h *Heap[T]) Len() int {
	return h.size
}

// Push adds an Element to the Heap.
func (h *Heap[T]) Push(e T) {
	// Increment first so that size points to the new position where we are
	// going to insert.
	h.size++

	// The initial order is log2(h.size)-1 because we want to simulate
	// running:
	//     level := h.size
	//     var stack Stack
	//     for level > 1 {
	//         stack.Push(level % 2)
	//         level /= 2
	//     }
	// (provided by the paper) backwards. By running it backwards, we can
	// avoid allocating our own stack and instead use recursion.
	h.root = h.insert(h.root, log2(h.size)-1, e)
}

func swapWithLeft[T Element[T]](cur T) T {
	oldRoot := cur
	newRoot := cur.Left()

	orr := oldRoot.Right()
	oldRoot.SetRight(newRoot.Right())
	newRoot.SetRight(orr)

	oldRoot.SetLeft(newRoot.Left())
	newRoot.SetLeft(oldRoot)
	return newRoot
}

func swapWithRight[T Element[T]](cur T) T {
	oldRoot := cur
	newRoot := cur.Right()

	orl := oldRoot.Left()
	oldRoot.SetLeft(newRoot.Left())
	newRoot.SetLeft(orl)

	oldRoot.SetRight(newRoot.Right())
	newRoot.SetRight(oldRoot)
	return newRoot
}

// insert is Push's recursive step. It fixes the Heap on the way back.
//
// insert assumes that the Heap's size has already been adjusted to account for
// the new Element.
func (h *Heap[T]) insert(cur T, order int, new T) T {
	if order < 0 {
		// Install the new leaf.
		var zero T
		new.SetLeft(zero)
		new.SetRight(zero)
		return new
	}

	// val = h.size / (2 ** order)
	val := h.size >> uint(order)
	if val&1 == 0 {
		// val is even, go left.
		cur.SetLeft(h.insert(cur.Left(), order-1, new))
		if cur.Left().Less(cur) {
			return swapWithLeft(cur)
		}
		return cur
	}

	// val is odd, go right.
	cur.SetRight(h.insert(cur.Right(), order-1, new))
	if cur.Right().Less(cur) {
		return swapWithRight(cur)
	}

	return cur
}

// Pop removes an Element from the Heap.
func (h *Heap[T]) Pop() T {
	if h.size == 0 {
		var zero T
		return zero
	}

	// Pull of the top Element and replace it with the bottom element.
	removed := h.root

	// See Push for why we start order at log2(h.size)-1.
	h.root = h.remove(h.root, log2(h.size)-1)

	// Decrement after so that size pointed to the position that we removed.
	h.size--

	// Fix the Heap.
	var zero T
	if h.root != zero {
		h.root.SetLeft(removed.Left())
		h.root.SetRight(removed.Right())
		h.root = h.fixDown(h.root)
	}

	return removed
}

// remove removes the last element from the heap and returns it.
//
// remove assumes that the Heap's size is the size before the removal.
func (h *Heap[T]) remove(cur T, order int) T {
	if order < 0 {
		// nil is a sentinel value. It means that this iteration hit the end
		// of the tree (or the tree was empty).
		var zero T
		return zero
	}

	// val = h.size / (2 ** order)
	val := h.size >> uint(order)

	if val&1 == 0 {
		// val is even, go left.
		got := h.remove(cur.Left(), order-1)
		var zero T
		if got == zero {
			// h.remove hit the end of the tree. Take the child.
			got = cur.Left()
			cur.SetLeft(zero)
		}
		return got
	}

	// val is odd, go right.
	got := h.remove(cur.Right(), order-1)
	var zero T
	if got == zero {
		// h.remove hit the end of the tree. Take the child.
		got = cur.Right()
		cur.SetRight(zero)
	}
	return got
}

// fixDown fixes a heap where only the root is potentially in the wrong place.
func (h *Heap[T]) fixDown(cur T) T {
	var zero T
	if cur.Left() == zero && cur.Right() == zero {
		return cur
	}

	// We have an "almost perfect" binary tree, so we now know that
	// cur.Left() != nil.
	if cur.Right() == zero || cur.Left().Less(cur.Right()) {
		// We only need to check the left child.
		if !cur.Left().Less(cur) {
			// Nothing to fix.
			return cur
		}
		newRoot := swapWithLeft(cur)
		newRoot.SetLeft(h.fixDown(newRoot.Left()))
		return newRoot
	}

	// We only need to check the right child.
	if !cur.Right().Less(cur) {
		// Nothing to fix.
		return cur
	}
	newRoot := swapWithRight(cur)
	newRoot.SetRight(h.fixDown(newRoot.Right()))
	return newRoot
}

type dummyElement[T Element[T]] struct {
	Entry[T]
}

func (e *dummyElement[T]) Less(elem T) bool {
	return true
}

// String implements fmt.Stringer.String.
func (h *Heap[T]) String() string {
	// Breadth first search.

	var l [][]T
	l = append(l, []T{h.root}, nil)

	var out string
	for len(l[0]) > 0 {
		f := l[0][0]
		l[0] = l[0][1:]

		out = fmt.Sprintf("%s%v", out, f)

		var zero T
		if f != zero {
			last := len(l) - 1
			l[last] = append(l[last], f.Left(), f.Right())
		}

		if len(l[0]) == 0 {
			l = l[1:]
			l = append(l, nil)
			out += "\n"
		} else {
			out += " "
		}
	}

	return out
}

// Entry is a default implementation of Linker. Users can embed this type in
// their structs to make them automatically implement most of the methods
// needed by Heap.
type Entry[T Element[T]] struct {
	left  T
	right T
}

// Left returns left child of e.
func (e *Entry[T]) Left() T {
	return e.left
}

// Right returns right child of e.
func (e *Entry[T]) Right() T {
	return e.right
}

// SetLeft assigns elem as the left child of e.
func (e *Entry[T]) SetLeft(elem T) {
	e.left = elem
}

// SetRight assigns elem as the right child of e.
func (e *Entry[T]) SetRight(elem T) {
	e.right = elem
}

// log2 calculates the integer log base 2 of n for positive n.
func log2(n int) int {
	if n <= 0 {
		panic(fmt.Sprint("log2 only defined on positive numbers, got ", n))
	}
	// Uses BSR or LZCNT on AMD64.
	return 63 - bits.LeadingZeros64(uint64(n))
}
