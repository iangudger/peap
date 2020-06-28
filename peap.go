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
// Recommended usage: Attach a little bow.
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
type Linker interface {
	Left() Element
	Right() Element
	SetLeft(Element)
	SetRight(Element)
}

// Element the item that is used at the API level.
//
// N.B. Like Linker, this is unlikely to be an interface in most cases. If
// Element is not an interface, it must be a pointer.
type Element interface {
	Linker
	Less(Element) bool
}

// Heap implement a pointer-based min-heap.
//
// +stateify savable
type Heap struct {
	size int
	root Element
}

// Peek returns the next Element to be removed from the Heap.
func (h *Heap) Peek() Element {
	return h.root
}

// Len returns the number of Elements currently in the Heap.
func (h *Heap) Len() int {
	return h.size
}

// Push adds an Element to the Heap.
func (h *Heap) Push(e Element) {
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

func swapWithLeft(cur Element) Element {
	oldRoot := cur
	newRoot := cur.Left()

	orr := oldRoot.Right()
	oldRoot.SetRight(newRoot.Right())
	newRoot.SetRight(orr)

	oldRoot.SetLeft(newRoot.Left())
	newRoot.SetLeft(oldRoot)
	return newRoot
}

func swapWithRight(cur Element) Element {
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
func (h *Heap) insert(cur Element, order int, new Element) Element {
	if order < 0 {
		// Install the new leaf.
		new.SetLeft(nil)
		new.SetRight(nil)
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
func (h *Heap) Pop() Element {
	if h.size == 0 {
		return nil
	}

	// Pull of the top Element and replace it with the bottom element.
	removed := h.root

	// See Push for why we start order at log2(h.size)-1.
	h.root = h.remove(h.root, log2(h.size)-1)

	// Decrement after so that size pointed to the position that we removed.
	h.size--

	// Fix the Heap.
	if h.root != nil {
		h.root.SetLeft(removed.Left())
		h.root.SetRight(removed.Right())
		h.root = h.fixDown(h.root)
	}

	return removed
}

// remove removes the last element from the heap and returns it.
//
// remove assumes that the Heap's size is the size before the removal.
func (h *Heap) remove(cur Element, order int) Element {
	if order < 0 {
		// nil is a sentinel value. It means that this iteration hit the end
		// of the tree (or the tree was empty).
		return nil
	}

	// val = h.size / (2 ** order)
	val := h.size >> uint(order)

	if val&1 == 0 {
		// val is even, go left.
		got := h.remove(cur.Left(), order-1)
		if got == nil {
			// h.remove hit the end of the tree. Take the child.
			got = cur.Left()
			cur.SetLeft(nil)
		}
		return got
	}

	// val is odd, go right.
	got := h.remove(cur.Right(), order-1)
	if got == nil {
		// h.remove hit the end of the tree. Take the child.
		got = cur.Right()
		cur.SetRight(nil)
	}
	return got
}

// fixDown fixes a heap where only the root is potentially in the wrong place.
func (h *Heap) fixDown(cur Element) Element {
	if cur.Left() == nil && cur.Right() == nil {
		return cur
	}

	// We have an "almost perfect" binary tree, so we now know that
	// cur.Left() != nil.
	if cur.Right() == nil || cur.Left().Less(cur.Right()) {
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

type dummyElement struct {
	Entry
}

func (e *dummyElement) Less(elem Element) bool {
	return true
}

// String implements fmt.Stringer.String.
func (h *Heap) String() string {
	// Breadth first search.

	var l []Element
	var end dummyElement
	l = append(l, h.root, &end)

	var out string
	for len(l) > 0 {
		f := l[0]
		l = l[1:]

		if f == nil {
			out += "nil "
			continue
		}
		if f.(*dummyElement) == &end {
			out += "\n"
			if len(l) > 0 {
				l = append(l, &end)
			}
			continue
		}

		out = fmt.Sprintf("%s%v ", out, f)
		l = append(l, f.Left(), f.Right())
	}

	return out
}

// Entry is a default implementation of Linker. Users can embed this type in
// their structs to make them automatically implement most of the methods
// needed by Heap.
//
// +stateify savable
type Entry struct {
	left  Element
	right Element
}

// Left returns left child of e.
func (e *Entry) Left() Element {
	return e.left
}

// Right returns right child of e.
func (e *Entry) Right() Element {
	return e.right
}

// SetLeft assigns elem as the left child of e.
func (e *Entry) SetLeft(elem Element) {
	e.left = elem
}

// SetRight assigns elem as the right child of e.
func (e *Entry) SetRight(elem Element) {
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
