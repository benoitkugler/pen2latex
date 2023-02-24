package symbols

import (
	"math/rand"
	"reflect"
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

func TestShape_normalizeY(t *testing.T) {
	tests := []struct {
		sh    Shape
		scope Rect
		want  Shape
	}{
		{
			Shape{{0, 0}, {0, 10}, {10, 10}},
			Rect{UL: Pos{0, 0}, LR: Pos{0, EMHeight}}, // no transformation
			Shape{{0, 0}, {0, 10}, {10, 10}},
		},
		{
			Shape{{0, 0}, {0, 10}, {10, 10}},
			Rect{UL: Pos{0, 0}, LR: Pos{0, EMHeight / 2}},
			Shape{{0, 0}, {0, 20}, {10, 20}},
		},
	}
	for _, tt := range tests {
		if got := tt.sh.normalizeY(tt.scope); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Shape.normalizeY() = %v, want %v", got, tt.want)
		}
	}
}
