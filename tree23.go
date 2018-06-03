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
// The tree has its own memory manager to avoid frequent allocations for single nodes that are created or removed.
package tree23

import (
    "errors"
    "fmt"
    "container/list"
)

type TreeElement interface {
    // The tree saves the corresponding max values of all children. To
    // achieve this, we have to get any kind of comparable number out of the tree elements.
    ExtractValue() float64
    // Calculates if two elements are equal
    Equal(e TreeElement) bool
}

type TreeNodeIndex int

type treeLink struct {
    maxChild float64
    child    TreeNodeIndex
}

type treeNode struct {
    // The slice can be nil for leaf nodes or contain a maximum of three
    // elements for maximum tree nodes with three children.
    children [3]treeLink
    cCount int
    // For all inner nodes, elem will be nil. Only leaf nodes contain a valid value!
    elem TreeElement
    // Links to the next or previous leaf node. This should build a continuous linked list
    // at the leaf level! Making range queries or iterations O(1).
    // Can only be expected to be valid for leaf nodes!
    prev TreeNodeIndex
    next TreeNodeIndex
}

// Tree23 is the exported tree type, that handles a complete tree structure!
// All caching, memory and data is handled inside this structure!
type Tree23 struct {

    // Root node access to the tree.
    root TreeNodeIndex

    // Caching of often used arrays/slices.
    g_oneElemTreeList   []TreeNodeIndex
    g_twoElemTreeList   []TreeNodeIndex
    g_threeElemTreeList []TreeNodeIndex
    g_nineElemTreeList  []TreeNodeIndex

    // Memory caching and node reusage.
    g_treeNodes []treeNode
    g_treeNodesFirstFreePos int
    g_treeNodesFreePositions *list.List
}

// Some global pre-allocated lists to avoid small allocations all the time.
func (tree *Tree23) initializeTree(capacity int) {

    tree.root = 0

    tree.g_oneElemTreeList   = []TreeNodeIndex{-1}
    tree.g_twoElemTreeList   = []TreeNodeIndex{-1, -1}
    tree.g_threeElemTreeList = []TreeNodeIndex{-1, -1, -1}
    tree.g_nineElemTreeList  = []TreeNodeIndex{-1, -1, -1, -1, -1, -1, -1, -1, -1}

    tree.g_treeNodes = make([]treeNode, capacity, capacity)
    for i:=0; i<len(tree.g_treeNodes); i++ {
        var a [3]treeLink
        tree.g_treeNodes[i] = treeNode{a, 0, nil, -1, -1}
    }
    tree.g_treeNodesFirstFreePos = 1
    tree.g_treeNodesFreePositions = list.New()
}

// NewCapacity Works exactly like New without paramters, but pre-allocated memory for the
// specified amount of maximum nodes beforehand. This may save some time for tree memory growing.
// If in doubt, use the normal New or provide a smaller number. The tree will not run out of memory!
func NewCapacity(expectedCapacity int) *Tree23 {

    var t Tree23
    t.initializeTree(expectedCapacity)
    return &t
}

// New creates a new tree that has no children and is not a leaf node!
// An empty tree from New can be used as base for inserting/deleting/searching.
// Runs in O(1)
func New() *Tree23 {
    return NewCapacity(1)
}

// IsLeaf returns true, if the given tree is a leaf node.
// Runs in O(1)
func (tree *Tree23) IsLeaf(t TreeNodeIndex) bool {
    return tree.g_treeNodes[t].cCount == 0
}

// IsEmpty returns true, if the given tree is empty (has no nodes)
// Runs in O(1)
func (tree *Tree23) IsEmpty(t TreeNodeIndex) bool {
    return tree.IsLeaf(t) && tree.g_treeNodes[t].elem == nil
}

// GetValue returns the value from a tree node.
// GetValue only works for leafs, as there is no data stored in other tree nodes!
// Please take care to only call GetValue on leaf nodes.
// Runs in O(1)
func (tree *Tree23) GetValue(t TreeNodeIndex) TreeElement {
    return tree.g_treeNodes[t].elem
}

// ChangeValue edits the value of a leaf node on the fly.
// ChangeValue only works for leafs, as there is no data stored in other tree nodes!
// Be very careful, to never edit properties, that may change the position in the tree!
// If the outcome of .ExtractValue() changes, the whole tree may become invalid beyond repair!
// Runs in O(1)
func (tree *Tree23) ChangeValue(t TreeNodeIndex, e TreeElement) {
    if tree.IsLeaf(t) {
        tree.g_treeNodes[t].elem = e
    }
}

// newNode returns a new node from cache or triggers a re-allocation for more memory!
func (tree *Tree23) newNode() TreeNodeIndex {

    // Recycle a deleted node.
    if tree.g_treeNodesFreePositions.Len() > 0 {
        e := tree.g_treeNodesFreePositions.Front()
        tree.g_treeNodesFreePositions.Remove(e)
        return e.Value.(TreeNodeIndex)
    }

    // Resize the cache and get more memory.
    // Resize our cache by 2x or 1.25x of the previous length. This is in accordance to slice append resizing.
    l := len(tree.g_treeNodes)
    if tree.g_treeNodesFirstFreePos >= l {
        appendSize := int(float64(l)*1.25)
        if l < 1000 {
            appendSize = l*2
        }
        tree.g_treeNodes = append(tree.g_treeNodes, make([]treeNode, appendSize)...)
    }

    // Get node from cached memory.
    tree.g_treeNodesFirstFreePos++
    return TreeNodeIndex(tree.g_treeNodesFirstFreePos-1)
}

// recycleNode adds the node into the stack for recycling. It will be reused when needed.
func (tree *Tree23) recycleNode(n TreeNodeIndex) {

    tree.g_treeNodes[n].cCount = 0
    tree.g_treeNodes[n].elem = nil
    tree.g_treeNodes[n].next = -1
    tree.g_treeNodes[n].prev = -1

    tree.g_treeNodesFreePositions.PushFront(n)
}

// newLeaf creates a new leaf node with an element and correct pointers.
func (tree *Tree23) newLeaf(elem TreeElement, prev, next TreeNodeIndex) TreeNodeIndex {

    n := tree.newNode()

    tree.g_treeNodes[n].cCount = 0
    tree.g_treeNodes[n].elem   = elem
    tree.g_treeNodes[n].prev   = prev
    tree.g_treeNodes[n].next   = next

    return n
}

// max returns the maximum element of the biggest subtree.
func (tree *Tree23) max(t TreeNodeIndex) float64 {
    if tree.IsLeaf(t) {
        return tree.g_treeNodes[t].elem.ExtractValue()
    }
    return tree.g_treeNodes[t].children[tree.g_treeNodes[t].cCount-1].maxChild
}

// nodeFromChildrenList creates a node from the list of children.
// The list can have a maximum of three children!
func (tree *Tree23) nodeFromChildrenList(children *[]TreeNodeIndex, startIndex, endIndex int) TreeNodeIndex {

    t := tree.newNode()
    tree.g_treeNodes[t].cCount = endIndex-startIndex

    index := 0
    for i := startIndex; i < endIndex; i++ {
        c := (*children)[i]
        tree.g_treeNodes[t].children[index] = treeLink{tree.max(c), c}
        index++
    }
    return t
}

// multipleNodesFromChildrenList returns between one and three nodes depending on the number of given children.
func (tree *Tree23) multipleNodesFromChildrenList(children *[]TreeNodeIndex, cLen int) *[]TreeNodeIndex {

    //cLen := len(*children)
    switch {
    case cLen <= 3:
        tree.g_oneElemTreeList[0] = tree.nodeFromChildrenList(children, 0, cLen)
        return &tree.g_oneElemTreeList
    case cLen <= 6:
        tree.g_twoElemTreeList[0] = tree.nodeFromChildrenList(children, 0, cLen/2)
        tree.g_twoElemTreeList[1] = tree.nodeFromChildrenList(children, cLen/2, cLen)
        return &tree.g_twoElemTreeList
    case cLen <= 9:
        tree.g_threeElemTreeList[0] = tree.nodeFromChildrenList(children, 0, cLen/3)
        tree.g_threeElemTreeList[1] = tree.nodeFromChildrenList(children, cLen/3, 2*cLen/3)
        tree.g_threeElemTreeList[2] = tree.nodeFromChildrenList(children, 2*cLen/3, cLen)
        return &tree.g_threeElemTreeList
    }
    // Should never get here!
    fmt.Println("SHOULD NOT GET HERE")

    return nil
}

// insertInto returns the first position bigger than the element itself or the last child to insert into!
func (tree *Tree23) insertInto(t TreeNodeIndex, elem TreeElement) int {

    v := elem.ExtractValue()
    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        // Find the tree with the smallest maximumChild bigger than elem itself!
        if v < tree.g_treeNodes[t].children[i].maxChild {
            return i
        }
    }

    return tree.g_treeNodes[t].cCount-1
}

// distributeTwoChildren creates a node with the two given children
func (tree *Tree23) distributeTwoChildren(c1, c2 TreeNodeIndex) TreeNodeIndex {

    n := tree.newNode()
    tree.g_treeNodes[n].cCount = 2

    tree.g_treeNodes[n].children[0].maxChild = tree.max(c1)
    tree.g_treeNodes[n].children[0].child = c1
    tree.g_treeNodes[n].children[1].maxChild = tree.max(c2)
    tree.g_treeNodes[n].children[1].child = c2
    return n
}

// distributeFourChildren creates a node with two sub-nodes with the four given children (two each)
func (tree *Tree23) distributeFourChildren(c1, c2, c3, c4 TreeNodeIndex) TreeNodeIndex {
    child1 := tree.distributeTwoChildren(c1, c2)
    child2 := tree.distributeTwoChildren(c3, c4)
    return tree.distributeTwoChildren(child1, child2)
}

// insertRec handles ecursive insertion. Returns a list of trees that are all on one level.
func (tree *Tree23) insertRec(t TreeNodeIndex, elem TreeElement) []TreeNodeIndex {

    if tree.IsLeaf(t) {

        if tree.g_treeNodes[t].elem.ExtractValue() < elem.ExtractValue() {
            leaf := tree.newLeaf(elem, t, tree.g_treeNodes[t].next)
            tree.g_treeNodes[t].next = leaf
            tree.g_treeNodes[tree.g_treeNodes[leaf].next].prev = leaf

            tree.g_twoElemTreeList[0] = t
            tree.g_twoElemTreeList[1] = leaf
            return tree.g_twoElemTreeList

        } else {
            leaf := tree.newLeaf(elem, tree.g_treeNodes[t].prev, t)
            tree.g_treeNodes[t].prev = leaf
            tree.g_treeNodes[tree.g_treeNodes[leaf].prev].next = leaf

            tree.g_twoElemTreeList[0] = leaf
            tree.g_twoElemTreeList[1] = t
            return tree.g_twoElemTreeList
        }

    }
    subTree := tree.insertInto(t, elem)
    // Recursive call to get a list of children back for redistribution :)
    // There can only ever be 1 or 2 children from the recursion!!!
    newChildren := tree.insertRec(tree.g_treeNodes[t].children[subTree].child, elem)

    // If we only get one child back, there is no re-ordering
    // necessary and the child can just be overwritten with the updated one.
    if len(newChildren) == 1 {

        tree.g_treeNodes[t].children[subTree].maxChild = tree.max(newChildren[0])
        tree.g_treeNodes[t].children[subTree].child = newChildren[0]

        return []TreeNodeIndex{t}
    }

    // Two children and two in our current tree. One of which is the updated
    // child coming from the recursion. So 3 in total. This is fine!
    // newChildren is already sorted! So we just have to figure out, where the new children go in our tree.
    // As newChildren should be within the bounds of [subTree] (smaller than the next node and bigger than the last)
    // we should replace the child at [subTree] and insert the second newChild directly afterwards.
    if tree.g_treeNodes[t].cCount == 2 {

        tree.g_treeNodes[t].children[subTree].maxChild = tree.max(newChildren[0])
        tree.g_treeNodes[t].children[subTree].child = newChildren[0]

        // We should move our second new child to index 1
        if subTree == 0 {
            tmpTreeNode := tree.g_treeNodes[t].children[1]
            tree.g_treeNodes[t].children[1] = treeLink{tree.max(newChildren[1]), newChildren[1]}
            tree.g_treeNodes[t].children[2] = tmpTreeNode
        } else {
            // We inserted into the second/last position and can just append our second new child.
            tree.g_treeNodes[t].children[2] = treeLink{tree.max(newChildren[1]), newChildren[1]}
        }
        tree.g_treeNodes[t].cCount = 3

        return []TreeNodeIndex{t}
    }

    defer tree.recycleNode(t)

    // We now have 3 original children (included [subTree]) and 2 new children from the recursion.
    // Both lists are separately sorted. And newChildren should fit perfectly into [subTree].
    // So we have to insert both newChildren at position subTree and should have a fully ordered tree!
    switch subTree {
    case 0:
        return []TreeNodeIndex{tree.distributeTwoChildren(newChildren[0], newChildren[1]),
            tree.distributeTwoChildren(tree.g_treeNodes[t].children[1].child, tree.g_treeNodes[t].children[2].child)}
    case 1:
        return []TreeNodeIndex{tree.distributeTwoChildren(tree.g_treeNodes[t].children[0].child, newChildren[0]),
            tree.distributeTwoChildren(newChildren[1], tree.g_treeNodes[t].children[2].child)}
    case 2:
        return []TreeNodeIndex{tree.distributeTwoChildren(tree.g_treeNodes[t].children[0].child, tree.g_treeNodes[t].children[1].child),
            tree.distributeTwoChildren(newChildren[0], newChildren[1])}
    }

    // We should never get here!
    return nil
}

// Insert inserts a given element into the tree.
// Runs in O(log(n))
func (tree *Tree23) Insert(elem TreeElement) {

    // This can only happen on an empty tree.
    if tree.IsEmpty(tree.root) {
        l := tree.newLeaf(elem, -1, -1)
        tree.g_treeNodes[l].prev = l
        tree.g_treeNodes[l].next = l
        tree.recycleNode(tree.root)
        tree.root = l
        return
    }

    // This can only happen on a tree with just one leaf.
    if tree.IsLeaf(tree.root) {

        l := tree.newLeaf(elem, -1, -1)

        if tree.g_treeNodes[l].elem.ExtractValue() < tree.g_treeNodes[tree.root].elem.ExtractValue() {
            tree.g_treeNodes[l].prev = tree.g_treeNodes[tree.root].prev
            tree.g_treeNodes[tree.g_treeNodes[l].prev].next = l
            tree.g_treeNodes[l].next = tree.root
            tree.g_treeNodes[tree.root].prev = l
            tree.root = tree.distributeTwoChildren(l, tree.root)
        } else {
            tree.g_treeNodes[l].prev = tree.root
            tree.g_treeNodes[l].next = tree.g_treeNodes[tree.root].next
            tree.g_treeNodes[tree.g_treeNodes[l].next].prev = l
            tree.g_treeNodes[tree.root].next = l
            tree.root = tree.distributeTwoChildren(tree.root, l)
        }
        return
    }

    subTree := tree.insertInto(tree.root, elem)
    newChildren := tree.insertRec(tree.g_treeNodes[tree.root].children[subTree].child, elem)

    // Returns a sorted tree (Rightfully replaces the node pointer)!
    if len(newChildren) == 1 {
        tree.g_treeNodes[tree.root].children[subTree].maxChild = tree.max(newChildren[0])
        tree.g_treeNodes[tree.root].children[subTree].child = newChildren[0]
        return
    }

    // We get two new children and have one old (subTree is overwritten!)
    if tree.g_treeNodes[tree.root].cCount == 2 {

        // Overwrite old child
        tree.g_treeNodes[tree.root].children[subTree].maxChild = tree.max(newChildren[0])
        tree.g_treeNodes[tree.root].children[subTree].child = newChildren[0]
        tree.g_treeNodes[tree.root].cCount = 3

        if subTree == 0 {
            tmpChild := tree.g_treeNodes[tree.root].children[1]
            tree.g_treeNodes[tree.root].children[1].maxChild = tree.max(newChildren[1])
            tree.g_treeNodes[tree.root].children[1].child = newChildren[1]
            tree.g_treeNodes[tree.root].children[2].maxChild = tree.max(tmpChild.child)
            tree.g_treeNodes[tree.root].children[2].child = tmpChild.child
        } else {
            tree.g_treeNodes[tree.root].children[2].maxChild = tree.max(newChildren[1])
            tree.g_treeNodes[tree.root].children[2].child = newChildren[1]
        }

        return
    }

    defer tree.recycleNode(tree.root)

    // We have 3 original children (one of which is at [subTree] and get another two newChildren
    switch subTree {
    case 0:
        tree.root = tree.distributeFourChildren(newChildren[0], newChildren[1], tree.g_treeNodes[tree.root].children[1].child, tree.g_treeNodes[tree.root].children[2].child)
    case 1:
        tree.root = tree.distributeFourChildren(tree.g_treeNodes[tree.root].children[0].child, newChildren[0], newChildren[1], tree.g_treeNodes[tree.root].children[2].child)
    case 2:
        tree.root = tree.distributeFourChildren(tree.g_treeNodes[tree.root].children[0].child, tree.g_treeNodes[tree.root].children[1].child, newChildren[0], newChildren[1])
    }

}

// deleteFrom returns the index of the child elem must be in (if any)
// It must the the first child bigger than elem itself. Or none.
// -1 is returned, if there exist no such child.
func (tree *Tree23) deleteFrom(t TreeNodeIndex, v float64) int {
    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        if v <= tree.g_treeNodes[t].children[i].maxChild {
            return i
        }
    }
    return -1
}

// deleteRec is the recursive function to delete elem in t.
// Returns a list of trees that are all on one level.
func (tree *Tree23) deleteRec(t TreeNodeIndex, elem TreeElement) *[]TreeNodeIndex {
    allLeaves := true

    leafCount := 0
    foundLeaf := false
    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        c := tree.g_treeNodes[t].children[i]
        isLeaf := tree.IsLeaf(c.child)
        allLeaves = allLeaves && isLeaf
        if isLeaf && (foundLeaf || !elem.Equal(tree.g_treeNodes[c.child].elem)) {
            leafCount++
        } else {
            // We only want to delete one node, that is equal to elem!
            // In case we successfully inserted multiple equal elements into our tree, we don't want to
            // remove all of them (tree can only handle -1 element at a time).
            foundLeaf = true
        }
    }
    if allLeaves {
        var newChildren *[]TreeNodeIndex

        // We cache the memory for this list!
        switch leafCount {
        case 1:
            newChildren = &tree.g_oneElemTreeList
        case 2:
            newChildren = &tree.g_twoElemTreeList
        case 3:
            newChildren = &tree.g_threeElemTreeList
        }

        index := 0
        foundLeaf = false
        for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
            c := tree.g_treeNodes[t].children[i]
            // Remove the child that contains our element!
            if foundLeaf || !elem.Equal(tree.g_treeNodes[c.child].elem) {
                (*newChildren)[index] = c.child
                index++
            } else {
                foundLeaf = true
                tree.g_treeNodes[tree.g_treeNodes[c.child].prev].next = tree.g_treeNodes[c.child].next
                tree.g_treeNodes[tree.g_treeNodes[c.child].next].prev = tree.g_treeNodes[c.child].prev

                tree.recycleNode(c.child)
            }
        }
        return newChildren
    }

    deleteFrom := tree.deleteFrom(t, elem.ExtractValue())
    // In case we don't find an element to delete, we just return our own children.
    // Let's get things sorted out some other recursion level.
    // No node recycling possible here.
    if deleteFrom == -1 {

        defer tree.recycleNode(t)

        switch leafCount {
        case 2:
            tree.g_twoElemTreeList[0] = tree.g_treeNodes[t].children[0].child
            tree.g_twoElemTreeList[1] = tree.g_treeNodes[t].children[1].child
            return &tree.g_twoElemTreeList
        case 3:
            tree.g_threeElemTreeList[0] = tree.g_treeNodes[t].children[0].child
            tree.g_threeElemTreeList[1] = tree.g_treeNodes[t].children[1].child
            tree.g_threeElemTreeList[2] = tree.g_treeNodes[t].children[2].child
            return &tree.g_threeElemTreeList
        }
    }

    // The new children from the subtree that does not contain elem any more!
    children := tree.deleteRec(tree.g_treeNodes[t].children[deleteFrom].child, elem)

    // Count the number of old grandChildren before allocating
    oGCCount := 0
    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        if i != deleteFrom {
            oGCCount += tree.g_treeNodes[tree.g_treeNodes[t].children[i].child].cCount
        }
    }

    // Includes all grandchildren and the new nodes from the recursion!
    index := 0

    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        c := tree.g_treeNodes[t].children[i]
        if i != deleteFrom {

            for j := 0; j < tree.g_treeNodes[c.child].cCount; j++ {
                c2 := tree.g_treeNodes[c.child].children[j]
                tree.g_nineElemTreeList[index] = c2.child
                index++
            }
        } else {
            // Here we insert the children from the recursion. They are now in sorted order with the rest!
            for _, c2 := range *children {
                tree.g_nineElemTreeList[index] = c2
                index++
            }
        }
        tree.recycleNode(c.child)
    }

    defer tree.recycleNode(t)

    return tree.multipleNodesFromChildrenList(&tree.g_nineElemTreeList, oGCCount+len(*children))
}

// Delete removes an element in the tree, if it exists. It will not throw any errors, if the element doesn't exist.
// Runs in O(log(n))
func (tree *Tree23) Delete(elem TreeElement) {

    if tree.IsEmpty(tree.root) {
        return
    }

    if tree.IsLeaf(tree.root) && elem.Equal(tree.g_treeNodes[tree.root].elem) {
        tree.g_treeNodes[tree.root].next = -1
        tree.g_treeNodes[tree.root].prev = -1
        tree.g_treeNodes[tree.root].elem = nil
        return
    }

    children := tree.deleteRec(tree.root, elem)

    defer tree.recycleNode(tree.root)

    if len(*children) == 1 {
        tree.root = (*children)[0]
        return
    }

    tree.root = tree.nodeFromChildrenList(children, 0, len(*children))
}

// findRec is the recursive function for finding elem in t.
// It returns the tree node (index) or an error if not found.
func (tree *Tree23) findRec(t TreeNodeIndex, elem TreeElement) (TreeNodeIndex, error) {
    if tree.IsLeaf(t) {
        if elem.Equal(tree.g_treeNodes[t].elem) {
            return t, nil
        } else {
            return -1, errors.New("TreeElement can not be found in the tree1.")
        }
    }

    subTree := tree.deleteFrom(t, elem.ExtractValue())
    if subTree == -1 {
        return -1, errors.New("TreeElement can not be found in the tree.")
    }

    return tree.findRec(tree.g_treeNodes[t].children[subTree].child, elem)
}

// Find tries to find the leaf node with the given element in t.
// If found, it will return the leaf node. Otherwise generated an error accordingly.
// Runs in O(log(n))
func (tree *Tree23) Find(elem TreeElement) (TreeNodeIndex, error) {
    if tree.IsEmpty(tree.root) {
        return -1, errors.New("Tree is empty. No elements can be found.")
    }
    return tree.findRec(tree.root, elem)
}

// findFirstLargerLeafRec is the recursive function for finding the smallest node bigger than value v in t.
func (tree *Tree23) findFirstLargerLeafRec(t TreeNodeIndex, v float64) (TreeNodeIndex, error) {
    if tree.IsLeaf(t) {
        if v <= tree.g_treeNodes[t].elem.ExtractValue() {
            return t, nil
        } else {
            return -1, errors.New("TreeElement can not be found in the tree.")
        }
    }

    subTree := tree.deleteFrom(t, v)
    if subTree == -1 {
        return -1, errors.New("TreeElement can not be found in the tree.")
    }

    return tree.findFirstLargerLeafRec(tree.g_treeNodes[t].children[subTree].child, v)
}

// FindFirstLargerLeaf returns the smallest leaf with a value bigger than v!
// If there is no such element, an error is returned ()
// Runs in O(log(n))
func (tree *Tree23) FindFirstLargerLeaf(v float64) (TreeNodeIndex, error) {
    if tree.IsEmpty(tree.root) {
        return -1, errors.New("Tree is empty. No elements can be found.")
    }

    return tree.findFirstLargerLeafRec(tree.root, v)
}

// Previous returns the previous leaf node that is smaller or equal than itself.
// For the smallest/first node in the tree, Previous will return the biggest/last node!
// Previous only works for leaf nodes and will generate an error otherwise.
// Runs in O(1)
func (tree *Tree23) Previous(t TreeNodeIndex) (TreeNodeIndex, error) {
    if tree.IsEmpty(t) {
        return -1, errors.New("Previous() does not work for empty trees")
    }
    if tree.IsLeaf(t) {
        return tree.g_treeNodes[t].prev, nil
    }
    return -1, errors.New("Previous() only works for leaf nodes!")
}

// Next returns the next leaf node that is bigger or equal than itself.
// For the biggest/last node in the tree, Next will return the smallest/first node!
// Next only works for leaf nodes and will generated an error otherwise.
// Runs in O(1)
func (tree *Tree23) Next(t TreeNodeIndex) (TreeNodeIndex, error) {
    if tree.IsEmpty(t) {
        return -1, errors.New("Next() does not work for empty trees")
    }
    if tree.IsLeaf(t) {
        return tree.g_treeNodes[t].next, nil
    }
    return -1, errors.New("Next() only works for leaf nodes!")
}

// minmaxDepth returns the minimum and maximum depth of all children (recursively) of t.
func (tree *Tree23) minmaxDepth(t TreeNodeIndex) (int, int) {
    if tree.IsEmpty(t) {
        return 0, 0
    }
    if tree.IsLeaf(t) {
        return 1, 1
    }
    depthMin := -1
    depthMax := -1

    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        c := tree.g_treeNodes[t].children[i]
        min, max := tree.minmaxDepth(c.child)
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
func (tree *Tree23) Depths() (int, int) {
    return tree.minmaxDepth(tree.root)
}

// getSmallestLeafRec is the recursive function that returns the left-most leaf node.
func (tree *Tree23) getSmallestLeafRec(t TreeNodeIndex) (TreeNodeIndex, error) {
    if tree.IsLeaf(t) {
        return t, nil
    }
    return tree.getSmallestLeafRec(tree.g_treeNodes[t].children[0].child)
}

// GetSmallestLeaf returns the leaf node of the smallest element in t
// or sets an error if the tree is empty.
// Runs in O(log(n))
func (tree *Tree23) GetSmallestLeaf() (TreeNodeIndex, error) {
    if tree.IsEmpty(tree.root) {
        return -1, errors.New("No leaf for an empty tree")
    }
    return tree.getSmallestLeafRec(tree.root)
}

// GetLargestLeaf returns the leaf node of the largest element in t
// or sets an error if the tree is empty.
// Runs in O(log(n))
func (tree *Tree23) GetLargestLeaf() (TreeNodeIndex, error) {
    l, err := tree.GetSmallestLeaf()
    if err != nil {
        return -1, err
    }

    return tree.Previous(l)
}

// checkLinkedList is the recursive function that runs through all leaf nodes by using
// the provided prev/next pointers and checks them on validity until it reaches the start node again.
func (tree *Tree23) checkLinkedList(startNode, currentNode TreeNodeIndex) bool {

    nextNode := tree.g_treeNodes[currentNode].next
    linkCheck := tree.g_treeNodes[nextNode].prev == currentNode

    // Once all around.
    if startNode == nextNode {
        return linkCheck
    }

    increasing := tree.g_treeNodes[nextNode].elem.ExtractValue() >= tree.g_treeNodes[currentNode].elem.ExtractValue()

    return linkCheck && increasing && tree.checkLinkedList(startNode, nextNode)
}

// leafListInvariant checks, that there are no dangling pointers and all elements are sorted increasingly!
func (tree *Tree23) leafListInvariant() bool {

    if tree.IsEmpty(tree.root) {
        return true
    }

    startNode, _ := tree.GetSmallestLeaf()
    return tree.checkLinkedList(startNode, startNode)
}

// Invariant checks the tree on validity.
// Returns true, if everything is OK with the given tree.
// Two things are checked: If the minimum and maximum depth is equal for every node up to the root.
// Further, the linked list for the leaf nodes is checked for valid increasing order and linking
// Including the link from the last to the first element.
// Runs in O(n)
func (tree *Tree23) Invariant() bool {
    depthMin, depthMax := tree.Depths()

    linkedListCorrect := tree.leafListInvariant()

    return depthMin == depthMax && linkedListCorrect
}

// pprint recursively pretty prints the tree.
func (tree *Tree23) pprint(t TreeNodeIndex, indentation int) {

    if tree.IsEmpty(t) {
        return
    }

    if tree.IsLeaf(t) {
        if indentation != 0 {
            fmt.Printf("  ")
        }
        for i := 0; i < indentation-1; i++ {
            fmt.Printf("|  ")
        }
        fmt.Printf("|")
        fmt.Printf("--(prev: %.2f. value: %.10f. next: %.2f)\n",
                tree.g_treeNodes[tree.g_treeNodes[t].prev].elem.ExtractValue(),
                tree.g_treeNodes[t].elem.ExtractValue(),
                tree.g_treeNodes[tree.g_treeNodes[t].next].elem.ExtractValue())
        return
    }

    for i := 0; i < tree.g_treeNodes[t].cCount; i++ {
        c := tree.g_treeNodes[t].children[i]
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
        tree.pprint(c.child, indentation + 1)
    }
}

// Pprint pretty prints the tree so it can be visually validated or understood.
// Runs in O(n log(n))
func (tree *Tree23) PrettyPrint() {
    tree.pprint(tree.root, 0)
    fmt.Printf("\n")
}
