package symbols

import (
	"math/rand"
	"testing"
)

func randPos() Pos {
	return Pos{
		X: rand.Float32(),
		Y: rand.Float32(),
	}
}

func randRect() Rect {
	return Rect{LR: randPos(), UL: randPos()}
}

func TestEmptyRect(t *testing.T) {
	for range [20]int{} {
		r := randRect()
		r2 := r
		r2.Union(EmptyRect())
		if r != r2 {
			t.Fatal()
		}
		empty := EmptyRect()
		empty.Union(r)
		if empty != r {
			t.Fatal()
		}
	}

	for range [20]int{} {
		empty := EmptyRect()
		p := randPos()
		empty.enlarge(p)
		if empty != (Rect{UL: p, LR: p}) {
			t.Fatal()
		}
	}
}
