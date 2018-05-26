package tree23

import (
	"testing"
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

func TestPreviousNext(t *testing.T) {
	tree := New()

	for i := 0; i <= 20; i++ {
		tree = tree.Insert(Element{i})
	}

	l, _ := tree.Find(Element{7})

	n, err := l.Next()
	if err != nil || n.elem.ExtractValue() <= 7 {
		t.Fail()
	}
	p, err := l.Previous()
	if err != nil || p.elem.ExtractValue() >= 7 {
		t.Fail()
	}
	if l2, err := p.Next(); err != nil || l2 != l {
		t.Fail()
	}
	if l2, err := n.Previous(); err != nil || l2 != l {
		t.Fail()
	}
	if !tree.Invariant() {
		t.Fail()
	}
}

func TestFindFirstLargerLeaf(t *testing.T) {
	tree := New()

	for i := 0; i <= 20; i++ {
		tree = tree.Insert(Element{i})
	}

	if e, err := tree.FindFirstLargerLeaf(3.5); err != nil || !e.elem.Equal(Element{4}) {
		t.Fail()
	}
	if e, err := tree.FindFirstLargerLeaf(-3.5); err != nil || !e.elem.Equal(Element{0}) {
		t.Fail()
	}
	if e, err := tree.FindFirstLargerLeaf(20.0); err != nil || !e.elem.Equal(Element{20}) {
		t.Fail()
	}
	if e, err := tree.FindFirstLargerLeaf(13.000001); err != nil || !e.elem.Equal(Element{14}) {
		t.Fail()
	}
	if e, err := tree.FindFirstLargerLeaf(13.999999); err != nil || !e.elem.Equal(Element{14}) {
		t.Fail()
	}
	if _, err := tree.FindFirstLargerLeaf(20.000001); err == nil {
		t.Fail()
	}

	if !tree.Invariant() {
		t.Fail()
	}
}

func TestFind(t *testing.T) {
	tree := New()

	for i := 0; i <= 20; i++ {
		tree = tree.Insert(Element{i})
	}

	if e, err := tree.Find(Element{13}); err != nil || !e.elem.Equal(Element{13}) {
		t.Fail()
	}
	if e, err := tree.Find(Element{7}); err != nil || !e.elem.Equal(Element{7}) {
		t.Fail()
	}
	if _, err := tree.Find(Element{23}); err == nil {
		t.Fail()
	}
	if _, err := tree.Find(Element{-2}); err == nil {
		t.Fail()
	}
	if !tree.Invariant() {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	tree := New()

	for i := 0; i < 100000; i++ {
		tree = tree.Insert(Element{i})
	}
	for i := 0; i < 100000; i++ {
		tree = tree.Delete(Element{i})
	}

	dMin, dMax := tree.Depths()

	if dMin != dMax || dMin != 0 || !tree.Invariant() || !tree.IsEmpty() {
		t.Fail()
	}

}

func TestInsert(t *testing.T) {
	tree := New()

	for i := 0; i < 40000; i++ {
		tree = tree.Insert(Element{i})
	}

	if !tree.Invariant() {
		t.Fail()
	}
}

func TestSmallestLargestLeaf(t *testing.T) {
	tree := New()

	for i := 0; i < 10000; i++ {
		tree = tree.Insert(Element{i})
	}

	if smallest, err := tree.GetSmallestLeaf(); err != nil || smallest.elem.(Element).E != 0 {
		t.Fail()
	}
	if largest, err := tree.GetLargestLeaf(); err != nil || largest.elem.(Element).E != 9999 {
		t.Fail()
	}
}

func TestNew(t *testing.T) {

	tree := New()
	if !tree.IsEmpty() {
		t.Fail()
	}
	if !tree.Invariant() {
		t.Fail()
	}

}
