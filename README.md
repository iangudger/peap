# Intrusive peap in Go

An implementation of a pointer-based (array-less) intrusive min-heap. Peap is a portmanteau of pointer heap.

An intrusive data structure requires elements to embed a part of the data structure. This improves performance by reducing allocations.

This peap is mostly for novelty purposes as it preforms worse than `container/heap` for all but very small heaps, even in the best possible case. If non-generic (interface) compatibility is dropped, the break even point can be pushed a bit higher, but for large heaps, an array-based heap will still be faster due to the cost of pointer chasing. Although it does release memory when no longer needed, adding array shrinking to `container/heap` in a wrapper would be simpler and likely preform better.

Compatible with `go_generics`.
