package tree23

import (
    "fmt"
    "testing"
    "time"
    "math/rand"
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

//func TestPreviousNext(t *testing.T) {
//    tree := New()
//
//    for i := 0; i <= 20; i++ {
//        tree.Insert(Element{i})
//    }
//
//    l, err := tree.Find(Element{7})
//
//    n, err := tree.Next(l)
//    if err != nil || tree.GetValue(n).ExtractValue() <= 7 {
//        t.Fail()
//    }
//    p, err := tree.Previous(l)
//    if err != nil || tree.GetValue(p).ExtractValue() >= 7 {
//        t.Fail()
//    }
//    if l2, err := tree.Next(p); err != nil || l2 != l {
//        t.Fail()
//    }
//    if l2, err := tree.Previous(n); err != nil || l2 != l {
//        t.Fail()
//    }
//    if !tree.Invariant() {
//        t.Fail()
//    }
//}
//
//func TestFindFirstLargerLeaf(t *testing.T) {
//    tree := New()
//
//    for i := 0; i <= 20; i++ {
//        tree.Insert(Element{i})
//    }
//
//    if e, err := tree.FindFirstLargerLeaf(3.5); err != nil || !tree.GetValue(e).Equal(Element{4}) {
//        t.Fail()
//    }
//    if e, err := tree.FindFirstLargerLeaf(-3.5); err != nil || !tree.GetValue(e).Equal(Element{0}) {
//        t.Fail()
//    }
//    if e, err := tree.FindFirstLargerLeaf(20.0); err != nil || !tree.GetValue(e).Equal(Element{20}) {
//        t.Fail()
//    }
//    if e, err := tree.FindFirstLargerLeaf(13.000001); err != nil || !tree.GetValue(e).Equal(Element{14}) {
//        t.Fail()
//    }
//    if e, err := tree.FindFirstLargerLeaf(13.999999); err != nil || !tree.GetValue(e).Equal(Element{14}) {
//        t.Fail()
//    }
//    if _, err := tree.FindFirstLargerLeaf(20.000001); err == nil {
//        t.Fail()
//    }
//
//    if !tree.Invariant() {
//        t.Fail()
//    }
//}
//
//func TestFind(t *testing.T) {
//    tree := New()
//
//    for i := 0; i <= 20; i++ {
//        tree.Insert(Element{i})
//    }
//
//    if e, err := tree.Find(Element{13}); err != nil || !tree.GetValue(e).Equal(Element{13}) {
//        t.Fail()
//    }
//    if e, err := tree.Find(Element{7}); err != nil || !tree.GetValue(e).Equal(Element{7}) {
//        t.Fail()
//    }
//    if _, err := tree.Find(Element{23}); err == nil {
//        t.Fail()
//    }
//    if _, err := tree.Find(Element{-2}); err == nil {
//        t.Fail()
//    }
//    if !tree.Invariant() {
//        t.Fail()
//    }
//}

//func TestInsert(t *testing.T) {
//    tree := New()
//
//    maxN := 1000000
//
//    for i := 0; i < maxN; i++ {
//        tree.Insert(Element{i})
//    }
//
//
//    for i := 0; i < maxN; i++ {
//        _,err := tree.Find(Element{i})
//        if err != nil {
//            fmt.Printf("Couldn't find %d\n", i)
//            t.Fail()
//        }
//    }
//
//    if !tree.Invariant() {
//        fmt.Printf("Invariant failed.\n")
//        t.Fail()
//    }
//}

//func TestDelete(t *testing.T) {
//
//    maxN := 1000000
//    tree := NewCapacity(maxN)
//
//    for i := 0; i < maxN; i++ {
//        tree.Insert(Element{i})
//    }
//    if !tree.Invariant() {
//        t.Fail()
//    }
//    for i := 0; i < maxN; i++ {
//        tree.Delete(Element{i})
//    }
//
//    dMin, dMax := tree.Depths()
//
//    if dMin != dMax || dMin != 0 || !tree.Invariant() {
//        t.Fail()
//    }
//
//}

func TestSimple(t *testing.T) {
    tree := New()

    maxN := 4

    var seed int64 = time.Now().UTC().UnixNano()
    seed = seed
    fmt.Printf("TestMemory Seed: %v\n", 1528022455179425997)
    r := rand.New(rand.NewSource(1528022455179425997))
    r = r

    for i := 0; i < maxN; i++ {
        tree.Insert(Element{i})
    }

    tree.PrettyPrint()

    if !tree.Invariant() {
        t.Fail()
    }

    tree.Delete(Element{1})

    tree.PrettyPrint()

    if !tree.Invariant() {
        t.Fail()
    }

}

//func TestMemory(t *testing.T) {
//    tree := New()
//
//    maxN := 10
//
//    for i := 0; i < maxN; i++ {
//        tree.Insert(Element{i})
//    }
//
//    var seed int64 = time.Now().UTC().UnixNano()
//    seed = seed
//    fmt.Printf("TestMemory Seed: %v\n", 1528022455179425997)
//    r := rand.New(rand.NewSource(1528022455179425997))
//
//    tree.PrettyPrint()
//
//    fmt.Println("Start memory check:")
//
//    // Run insert and delete a lot!! Theoretically, we should be able
//    // to recycle all nodes that are removed/added and should not hit any limits!
//    for i := 0; i < 1000; i++ {
//
//        randInt := r.Intn(maxN)
//
//        _,err := tree.Find(Element{randInt})
//        if err != nil {
//            fmt.Printf("Cannot find element %d, err == %v\n", randInt, err)
//        }
//
//        tree.Delete(Element{randInt})
//        tree.Insert(Element{randInt})
//
//        _,err = tree.Find(Element{randInt})
//        if err != nil {
//            fmt.Printf("Cannot find element %d, err == %v\n", randInt, err)
//        }
//
//    }
//
//    if !tree.Invariant() {
//        t.Fail()
//    }
//    for i := 0; i < maxN; i++ {
//        tree.Delete(Element{i})
//    }
//
//    dMin, dMax := tree.Depths()
//
//    if dMin != dMax || dMin != 0 || !tree.Invariant() {
//        t.Fail()
//    }
//
//}

//func TestSmallestLargestLeaf(t *testing.T) {
//    tree := New()
//
//    maxN := 1000000
//
//    for i := 0; i < maxN; i++ {
//        tree.Insert(Element{i})
//    }
//
//    if smallest, err := tree.GetSmallestLeaf(); err != nil || tree.GetValue(smallest).(Element).E != 0 {
//        t.Fail()
//    }
//    if largest, err := tree.GetLargestLeaf(); err != nil || tree.GetValue(largest).(Element).E != maxN-1 {
//        t.Fail()
//    }
//    if !tree.Invariant() {
//        t.Fail()
//    }
//}
