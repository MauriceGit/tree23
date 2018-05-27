// MIT License
//
// Copyright (c) 2018 Maurice Tollmien (maurice.tollmien@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// tree23 is an implementation for a balanced 2-3-tree.
// It distinguishes itself from other implementations of 2-3-trees by having a few more
// functions defined for finding elements close to a key (similar to possible insert positions in the tree)
// for floating point keys and by having a native function to retreive the next or previous leaf element
// in the tree without knowing its key or position in the tree that work in O(1) for every leaf!
// The last element links to the first and the first back to the last element.
package tree23

import (
    "errors"
    "fmt"
)

type TreeElement interface {
    // The tree saves the corresponding max values of all children. To
    // achieve this, we have to get any kind of comparable number out of the tree elements.
    ExtractValue() float64
    // Calculates if two elements are equal
    Equal(e TreeElement) bool
}

type TreeNode struct {
    maxChild float64
    child    *Tree23
}

type Tree23 struct {
    // The slice can be nil for leaf nodes or contain a maximum of three
    // elements for maximum tree nodes with three children.
    children []TreeNode
    // For all inner nodes, elem will be nil. Only leaf nodes contain a valid value!
    elem TreeElement
    // Links to the next or previous leaf node. This should build a continuous linked list
    // at the leaf level! Making range queries or iterations O(1).
    // Can only be expected to be valid for leaf nodes!
    prev *Tree23
    next *Tree23
}

type TreeList []*Tree23

// Some global pre-allocated lists to avoid small allocations all the time.
var g_oneElemTreeList []*Tree23 = []*Tree23{nil}
var g_twoElemTreeList []*Tree23 = []*Tree23{nil, nil}
var g_threeElemTreeList []*Tree23 = []*Tree23{nil, nil, nil}

// New creates a new tree that has no children and is not a leaf node!
// An empty tree from New can be used as base for inserting/deleting/searching.
// Runs in O(1)
func New() *Tree23 {
    return &Tree23{make([]TreeNode, 0), nil, nil, nil}
}

// IsLeaf returns true, if the given tree is a leaf node.
// Runs in O(1)
func (t *Tree23) IsLeaf() bool {
    return len(t.children) == 0
}

// IsEmpty returns true, if the given tree is empty (has no nodes)
// Runs in O(1)
func (t *Tree23) IsEmpty() bool {
    return t.IsLeaf() && t.elem == nil
}

// GetValue returns the value from a tree node.
// GetValue only works for leafs, as there is no data stored in other tree nodes!
// Runs in O(1)
func (t *Tree23) GetValue() (TreeElement, error) {
    if t.IsLeaf() {
        return t.elem, nil
    }
    return nil, errors.New("No value available for non-leaf nodes")
}

// GetValueEditable returns the pointer to a value from a tree node.
// GetValueEditable only works for leafs, as there is no data stored in other tree nodes!
// Be very careful, to never edit properties, that may change the position in the tree!
// If the outcome of .ExtractValue() changes, the whole tree may become invalid beyond repair!
// Runs in O(1)
//func (t *Tree23) GetValueEditable() (*TreeElement, error) {
//    if t.IsLeaf() {
//        return &t.elem, nil
//    }
//    return nil, errors.New("No value available for non-leaf nodes")
//}

// ChangeValue edits the value of a leaf node on the fly.
// ChangeValue only works for leafs, as there is no data stored in other tree nodes!
// Be very careful, to never edit properties, that may change the position in the tree!
// If the outcome of .ExtractValue() changes, the whole tree may become invalid beyond repair!
// Runs in O(1)
func (t *Tree23) ChangeValue(e TreeElement) {
    if t.IsLeaf() {
        t.elem = e
    }
}

// Keeps a sorted tree sorted.
func newLeaf(elem TreeElement, prev, next *Tree23) *Tree23 {
    return &Tree23{make([]TreeNode, 0), elem, prev, next}
}

// Returns the maximum element in the subtree.
func (t *Tree23) max() float64 {
    if t.IsLeaf() {
        return t.elem.ExtractValue()
    }
    return t.children[len(t.children)-1].maxChild
}

// Creates a node from the list of children.
// The list can have a maximum of three children!
func nodeFromChildrenList(children *[]*Tree23, startIndex, endIndex int) *Tree23 {
    t := &Tree23{make([]TreeNode, endIndex-startIndex), nil, nil, nil}

    index := 0
    for i := startIndex; i < endIndex; i++ {
        c := (*children)[i]
        t.children[index] = TreeNode{c.max(), c}
        index++
    }
    return t
}

// Returns between one and three nodes depending on the number of given children.
func multipleNodesFromChildrenList(children []*Tree23) []*Tree23 {

    cLen := len(children)
    switch {
    case cLen <= 3:
        g_oneElemTreeList[0] = nodeFromChildrenList(&children, 0, cLen)
        return g_oneElemTreeList
    case cLen <= 6:
        g_twoElemTreeList[0] = nodeFromChildrenList(&children, 0, cLen/2)
        g_twoElemTreeList[1] = nodeFromChildrenList(&children, cLen/2, cLen)
        return g_twoElemTreeList
    case cLen <= 9:
        g_threeElemTreeList[0] = nodeFromChildrenList(&children, 0, cLen/3)
        g_threeElemTreeList[1] = nodeFromChildrenList(&children, cLen/3, 2*cLen/3)
        g_threeElemTreeList[2] = nodeFromChildrenList(&children, 2*cLen/3, cLen)
        return g_threeElemTreeList
    }
    // Should never get here!
    return []*Tree23{}
}

// Returns the first position bigger than the element itself or the last child to insert into!
func (t *Tree23) insertInto(elem TreeElement) int {

    for i, c := range t.children {
        // Find the tree with the smallest maximumChild bigger than elem itself!
        if elem.ExtractValue() < c.maxChild {
            return i
        }
    }

    return len(t.children) - 1
}

// Creates a node with the two given children
func distributeTwoChildren(c1, c2 *Tree23) *Tree23 {
    child := &Tree23{make([]TreeNode, 2), nil, nil, nil}
    child.children[0].maxChild = c1.max()
    child.children[0].child = c1
    child.children[1].maxChild = c2.max()
    child.children[1].child = c2
    return child
}

// Creates a node with two sub-nodes with the four given children (two each)
func distributeFourChildren(c1, c2, c3, c4 *Tree23) *Tree23 {
    child1 := distributeTwoChildren(c1, c2)
    child2 := distributeTwoChildren(c3, c4)
    return distributeTwoChildren(child1, child2)
}

// Recursive insertion. Returns a list of trees on one level.
func (t *Tree23) insertRec(elem TreeElement) []*Tree23 {

    if t.IsLeaf() {

        if t.elem.ExtractValue() < elem.ExtractValue() {
            leaf := newLeaf(elem, t, t.next)
            t.next = leaf
            leaf.next.prev = leaf

            g_twoElemTreeList[0] = t
            g_twoElemTreeList[1] = leaf
            return g_twoElemTreeList

        } else {
            leaf := newLeaf(elem, t.prev, t)
            t.prev = leaf
            leaf.prev.next = leaf

            g_twoElemTreeList[0] = leaf
            g_twoElemTreeList[1] = t
            return g_twoElemTreeList
        }

    }
    subTree := t.insertInto(elem)
    // Recursive call to get a list of children back for redistribution :)
    // There can only ever be 1 or 2 children from the recursion!!!
    newChildren := t.children[subTree].child.insertRec(elem)

    // If we only get one child back, there is no re-ordering
    // necessary and the child can just be overwritten with the updated one.
    if len(newChildren) == 1 {
        t.children[subTree].maxChild = newChildren[0].max()
        t.children[subTree].child = newChildren[0]
        return []*Tree23{t}
    }

    // Two children and two in our current tree. One of which is the updated
    // child coming from the recursion. So 3 in total. This is fine!
    // newChildren is already sorted! So we just have to figure out, where the new children go in our tree.
    // As newChildren should be within the bounds of [subTree] (smaller than the next node and bigger than the last)
    // we should replace the child at [subTree] and insert the second newChild directly afterwards.
    if len(t.children) == 2 {
        t.children[subTree].maxChild = newChildren[0].max()
        t.children[subTree].child = newChildren[0]

        // We should move our second new child to index 1
        if subTree == 0 {
            tmpTreeNode := t.children[1]
            t.children[1] = TreeNode{newChildren[1].max(), newChildren[1]}
            t.children = append(t.children, tmpTreeNode)
        } else {
            // We inserted into the second/last position and can just append our second new child.
            t.children = append(t.children, TreeNode{newChildren[1].max(), newChildren[1]})
        }

        return []*Tree23{t}
    }

    // We now have 3 original children (included [subTree]) and 2 new children from the recursion.
    // Both lists are separately sorted. And newChildren should fit perfectly into [subTree].
    // So we have to insert both newChildren at position subTree and should have a fully ordered tree!
    switch subTree {
    case 0:
        return []*Tree23{distributeTwoChildren(newChildren[0], newChildren[0]),
            distributeTwoChildren(t.children[1].child, t.children[2].child)}
    case 1:
        return []*Tree23{distributeTwoChildren(t.children[0].child, newChildren[0]),
            distributeTwoChildren(newChildren[1], t.children[2].child)}
    case 2:
        return []*Tree23{distributeTwoChildren(t.children[0].child, t.children[1].child),
            distributeTwoChildren(newChildren[0], newChildren[1])}
    }

    // We should never get here!
    return nil
}

// Insert inserts a given element into the tree.
// Returns a new instance of the tree!
// The root node may change, so reassign the tree to the output of this function!
// Runs in O(log(n))
func (t *Tree23) Insert(elem TreeElement) *Tree23 {

    // This can only happen on an empty tree.
    if t.IsEmpty() {
        l := newLeaf(elem, nil, nil)
        l.prev = l
        l.next = l
        return l
    }

    // This can only happen on a tree with just one leaf.
    if t.IsLeaf() {

        newChild := newLeaf(elem, nil, nil)

        if newChild.elem.ExtractValue() < t.elem.ExtractValue() {
            newChild.prev = t.prev
            newChild.prev.next = newChild
            newChild.next = t
            t.prev = newChild
            return distributeTwoChildren(newChild, t)
        } else {
            newChild.prev = t
            newChild.next = t.next
            newChild.next.prev = newChild
            t.next = newChild
            return distributeTwoChildren(t, newChild)
        }
    }

    subTree := t.insertInto(elem)
    newChildren := t.children[subTree].child.insertRec(elem)

    // Returns a sorted tree (Rightfully replaces the node pointer)!
    if len(newChildren) == 1 {
        t.children[subTree].maxChild = newChildren[0].max()
        t.children[subTree].child = newChildren[0]
        return t
    }

    // We get two new children and have one old (subTree is overwritten!)
    if len(t.children) == 2 {
        // Overwrite old child
        t.children[subTree].maxChild = newChildren[0].max()
        t.children[subTree].child = newChildren[0]

        if subTree == 0 {
            tmpChild := t.children[1]
            t.children[1].maxChild = newChildren[1].max()
            t.children[1].child = newChildren[1]
            t.children = append(t.children, TreeNode{tmpChild.child.max(), tmpChild.child})
        } else {
            t.children = append(t.children, TreeNode{newChildren[1].max(), newChildren[1]})
        }

        return t
    }

    // We have 3 original children (one of which is at [subTree] and get another two newChildren
    switch subTree {
    case 0:
        return distributeFourChildren(newChildren[0], newChildren[0], t.children[1].child, t.children[2].child)
    case 1:
        return distributeFourChildren(t.children[0].child, newChildren[0], newChildren[1], t.children[2].child)
    case 2:
        return distributeFourChildren(t.children[0].child, t.children[1].child, newChildren[0], newChildren[1])
    }

    // We should never get here!
    return nil
}

// Returns the index of the child elem must be in (if any)
// It must the the first child bigger than elem itself. Or none.
// -1 is returned, if there exist no such child.
func (t *Tree23) deleteFrom(v float64) int {
    for i, _ := range t.children {
        if v <= t.children[i].maxChild {
            return i
        }
    }
    return -1
}

// Recursive function to delete elem in t.
// Returns a list of trees on one level.
func (t *Tree23) deleteRec(elem TreeElement) []*Tree23 {
    allLeaves := true

    leafCount := 0
    for _, c := range t.children {
        IsLeaf := c.child.IsLeaf()
        allLeaves = allLeaves && IsLeaf
        if IsLeaf && !elem.Equal(c.child.elem) {
            leafCount++
        }
    }
    if allLeaves {
        var newChildren []*Tree23

        // We cache the memory for this list!
        switch leafCount {
        case 1:
            newChildren = g_oneElemTreeList
        case 2:
            newChildren = g_twoElemTreeList
        case 3:
            newChildren = g_threeElemTreeList
        }

        index := 0
        for _, c := range t.children {
            // Remove the child that contains our element!
            if !elem.Equal(c.child.elem) {
                newChildren[index] = c.child
                index++
            } else {
                c.child.prev.next = c.child.next
                c.child.next.prev = c.child.prev
            }
        }
        return newChildren
    }

    deleteFrom := t.deleteFrom(elem.ExtractValue())
    // In case we don't find an element to delete, we just return our own children.
    // Let's get things sorted out some other recursion level.
    if deleteFrom == -1 {
        var children []*Tree23
        switch leafCount {
        case 2:
            children = g_twoElemTreeList
        case 3:
            children = g_threeElemTreeList
        }

        for i, c := range t.children {
            children[i] = c.child
        }
        return children
    }

    // The new children from the subtree that does not contain elem any more!
    children := t.children[deleteFrom].child.deleteRec(elem)

    // Count the number of old grandChildren before allocating
    oGCCount := 0
    for i, c := range t.children {
        if i != deleteFrom {
            oGCCount += len(c.child.children)
        }
    }

    // Includes all grandchildren and the new nodes from the recursion!
    allNodes := make([]*Tree23, oGCCount+len(children))
    index := 0

    for i, c := range t.children {
        if i != deleteFrom {
            for _, c2 := range c.child.children {
                allNodes[index] = c2.child
                index++
            }
        } else {
            // Here we insert the children from the recursion. They are now in sorted order with the rest!
            for _, c2 := range children {
                allNodes[index] = c2
                index++
            }
        }
    }

    return multipleNodesFromChildrenList(allNodes)
}

// Delete removes an element in the tree, if it exists.
// Returns a new instance of the tree!
// The root node may change, so reassign the tree to the output of this function!
// Runs in O(log(n))
func (t *Tree23) Delete(elem TreeElement) *Tree23 {
    if t.IsEmpty() || t.IsLeaf() && elem.Equal(t.elem) {
        return &Tree23{make([]TreeNode, 0), nil, nil, nil}
    }

    children := t.deleteRec(elem)

    if len(children) == 1 {
        return children[0]
    }

    return nodeFromChildrenList(&children, 0, len(children))
}

func (t *Tree23) findRec(elem TreeElement) (*Tree23, error) {
    if t.IsLeaf() {
        if elem.Equal(t.elem) {
            return t, nil
        } else {
            return nil, errors.New("TreeElement can not be found in the tree.")
        }
    }

    subTree := t.deleteFrom(elem.ExtractValue())
    if subTree == -1 {
        return nil, errors.New("TreeElement can not be found in the tree.")
    }

    return t.children[subTree].child.findRec(elem)
}

// Find tries to find the leaf node with the given element in t.
// If found, it will return the leaf node. Otherwise generated an error accordingly.
// Runs in O(log(n))
func (t *Tree23) Find(elem TreeElement) (*Tree23, error) {
    if t.IsEmpty() {
        return nil, errors.New("Tree is empty. No elements can be found.")
    }
    return t.findRec(elem)
}

func (t *Tree23) findFirstLargerLeafRec(v float64) (*Tree23, error) {
    if t.IsLeaf() {
        if v <= t.elem.ExtractValue() {
            return t, nil
        } else {
            return nil, errors.New("TreeElement can not be found in the tree.")
        }
    }

    subTree := t.deleteFrom(v)
    if subTree == -1 {
        return nil, errors.New("TreeElement can not be found in the tree.")
    }

    return t.children[subTree].child.findFirstLargerLeafRec(v)
}

// FindFirstLargerLeaf returns the smallest leaf with a value bigger than v!
// If there is no such element, an error is returned ()
// Runs in O(log(n))
func (t *Tree23) FindFirstLargerLeaf(v float64) (*Tree23, error) {
    if t.IsEmpty() {
        return nil, errors.New("Tree is empty. No elements can be found.")
    }

    return t.findFirstLargerLeafRec(v)
}

// Next returns the next leaf node that is bigger than itself.
// For the biggest/last node in the tree, Next will return the smallest/first node!
// Previous only works for leaf nodes and will generate an error otherwise.
// Runs in O(1)
func (t *Tree23) Previous() (*Tree23, error) {
    if t.IsEmpty() {
        return nil, errors.New("Previous() does not work for empty trees")
    }
    if t.IsLeaf() {
        return t.prev, nil
    }
    return nil, errors.New("Previous() only works for leaf nodes!")
}

// Next returns the next leaf node that is bigger than itself.
// For the biggest/last node in the tree, Next will return the smallest/first node!
// Next only works for leaf nodes and will generated an error otherwise.
// Runs in O(1)
func (t *Tree23) Next() (*Tree23, error) {
    if t.IsEmpty() {
        return nil, errors.New("Next() does not work for empty trees")
    }
    if t.IsLeaf() {
        return t.next, nil
    }
    return nil, errors.New("Next() only works for leaf nodes!")
}

func (t *Tree23) minmaxDepth() (int, int) {
    if t.IsEmpty() {
        return 0, 0
    }
    if t.IsLeaf() {
        return 1, 1
    }
    depthMin := -1
    depthMax := -1

    for _, c := range t.children {
        min, max := c.child.minmaxDepth()
        if depthMin == -1 || min < depthMin {
            depthMin = min + 1
        }
        if depthMax == -1 || max > depthMax {
            depthMax = max + 1
        }
    }
    return depthMin, depthMax
}

// Depths returns the minimum and maximum depth of the tree t.
// minimum and maximum should always be the same ()
// Runs in O(log(n))
func (t *Tree23) Depths() (int, int) {
    return t.minmaxDepth()
}

func (t *Tree23) getSmallestLeafRec() (*Tree23, error) {
    if t.IsLeaf() {
        return t, nil
    }
    return t.children[0].child.getSmallestLeafRec()
}

// GetSmallestLeaf returns the leaf node of the smallest element in t
// or sets an error if the tree is empty.
// Runs in O(log(n))
func (t *Tree23) GetSmallestLeaf() (*Tree23, error) {
    if t.IsEmpty() {
        return nil, errors.New("No leaf for an empty tree")
    }
    return t.getSmallestLeafRec()
}

// GetLargestLeaf returns the leaf node of the largest element in t
// or sets an error if the tree is empty.
// Runs in O(log(n))
func (t *Tree23) GetLargestLeaf() (*Tree23, error) {
    l, err := t.GetSmallestLeaf()
    if err != nil {
        return nil, err
    }

    return l.Previous()
}

func checkLinkedList(startNode, currentNode *Tree23) bool {

    nextNode := currentNode.next
    linkCheck := nextNode.prev == currentNode

    // Once all around.
    if startNode.elem.Equal(nextNode.elem) {
        return linkCheck
    }

    increasing := nextNode.elem.ExtractValue() > currentNode.elem.ExtractValue()

    return linkCheck && increasing && checkLinkedList(startNode, nextNode)
}

// Checks, that there are no dangling pointers and all elements are sorted increasingly!
func (t *Tree23) leafListInvariant() bool {

    if t.IsEmpty() {
        return true
    }

    startNode, _ := t.GetSmallestLeaf()
    return checkLinkedList(startNode, startNode)
}

// Invariant checks the tree on validity.
// Returns true, if everything is OK with the given tree.
// Two things are checked: If the minimum and maximum depth is equal for every node up to the root.
// Further, the linked list for the leaf nodes is checked for valid increasing order and linking
// Including the link from the last to the first element.
// Runs in O(n)
func (t *Tree23) Invariant() bool {
    depthMin, depthMax := t.Depths()

    linkedListCorrect := t.leafListInvariant()

    return depthMin == depthMax && linkedListCorrect
}

func (t *Tree23) pprint(indentation int) {

    if t.IsEmpty() {
        return
    }

    if t.IsLeaf() {
        if indentation != 0 {
            fmt.Printf("  ")
        }
        for i := 0; i < indentation-1; i++ {
            fmt.Printf("|  ")
        }
        fmt.Printf("|")
        fmt.Printf("--(prev: %.0f. value: %.0f. next: %.0f)\n", t.prev.elem.ExtractValue(), t.elem.ExtractValue(), t.next.elem.ExtractValue())
        return
    }

    for _, c := range t.children {
        if indentation != 0 {
            fmt.Printf("  ")
        }
        for i := 0; i < indentation-1; i++ {
            fmt.Printf("|  ")
        }
        if indentation != 0 {
            fmt.Printf("|")
        }
        fmt.Printf("--%.0f\n", c.maxChild)
        c.child.pprint(indentation + 1)
    }
}

// String overwrites the standard string routine for use in Pprint().
func (s TreeList) String() string {
    st := "["
    for _, c := range s {
        st += fmt.Sprintf("%.0f, ", c.elem.ExtractValue())
    }
    st += fmt.Sprintf("]\n")
    return st
}

// Pprint pretty prints the tree so it can be visually validated or understood.
// Runs in O(n log(n))
func (t *Tree23) Pprint() {
    t.pprint(0)
    fmt.Printf("\n")
}
