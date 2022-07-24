# Intrusive peap in Go

An implementation of a pointer-based (array-less) intrusive min-heap. Peap is a portmanteau of pointer heap.

An intrusive data structure requires elements to embed a part of the data structure. This improves performance by reducing allocations.

This peap is mostly for novelty purposes as it preforms worse than `container/heap` for all but very small heaps, even in the best possible case. An array-based heap will be faster due to the cost of pointer chasing. Even with `container/heap` doing interface conversions and array doubling, it amortizes to being faster at somewhere fewer than 10 elements in the heap. Although it does release memory when no longer needed, adding array shrinking to `container/heap` in a wrapper would be simpler and likely preform better.

