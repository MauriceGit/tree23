# 2,3-Tree

A Go library that implements a completely balanced 2,3-Tree that allows generic data as values in the tree.
A 2,3-Tree self-balances itself when inserting or deleting elements and it is guaranteed, that all leaf nodes will
be at a level at all times. The balancing itself runs in O(1) while ascending the tree.
All nodes will be positioned at the leaf-level. Inserting/Deleting/Searching all take O(log n) time with the logarithm between base 2 and 3.

Additionally, all leaf nodes are linked to another in a sorted order (and circularly). This additionally allows accessing
previous or next items (when we already have a leaf node) in O(1) without the need to further traverse the tree.
An iterator is simply achieved by retrieving the smallest child (there is a function for that) and following the .Next() links.

Further, the tree allows accessing elements close to a given value (without knowing the exact leaf value!).

## Documentation:

For detailed functionality and API, please have a look at godoc.

A fully functional example of creating a tree, inserting/deleting/searching:

```go

import (
    "github.com/MauriceGit/tree23"
)

type Element struct {
    E int
}

func (e Element) Equal(e2 TreeElement) bool {
    return e.E == e2.(Element).E
}
func (e Element) ExtractValue() float64 {
    return float64(e.E)
}

func main() {
    tree := New()

    for i := 0; i < 100000; i++ {
        tree = tree.Insert(Element{i})
    }

    // e1 will be the leaf node: 14
    e1, err := tree.FindFirstLargerLeaf(13.000001)
    // e2 will be the leaf node: 7
    e2, err := tree.Find(Element{7})
    // e3 will be the leaf node: 8
    e3, err := e2.Next()
    // e4 will be the leaf node: 6
    e4, err := e2.Previous()
    // smallest will be the leaf node: 0
    smallest, err := tree.GetSmallestLeaf()
    // largest will be the leaf node: 99999
    largest, err  := tree.GetLargestLeaf()

    for i := 0; i < 100000; i++ {
        tree = tree.Delete(Element{i})
    }

    // tree is now empty.

    ...
}

```
