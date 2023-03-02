package symbols

import (
	"reflect"
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

func Test_frechetDistance(t *testing.T) {
	u := Shape{{0, 0}, {1, 0}, {1, 1}, {1, 2}}
	v := u
	if got := frechetDistanceShapes(u, v); got != 0 {
		t.Errorf("frechetDistance() = %v, want %v", got, 0)
	}

	w := Shape{{0, 0}, {1, 4}, {1, 1}, {1, 2}}
	if got := frechetDistanceShapes(u, w); got <= 0 {
		t.Errorf("frechetDistance() = %v, want > 0", got)
	}
}

func Test_closestPointDistance(t *testing.T) {
	tests := []struct {
		u    Shape
		v    Shape
		want fl
	}{
		{
			Shape{{0, 10}, {0, 20}, {0, 30}}, Shape{{0, 11}}, 1,
		},
		{
			Shape{{0, 10}, {0, 20}, {0, 30}}, Shape{{0, 12}, {0, 21}}, 1,
		},
	}
	for _, tt := range tests {
		if got := closestPointDistance(tt.u, tt.v); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("closestPointDistance() = %v, want %v", got, tt.want)
		}
	}
}

func TestSegment_distance(t *testing.T) {
	tests := []struct {
		U     Segment
		V     Segment
		want  trans
		want1 fl
	}{
		{Segment{Pos{}, Pos{10, 0}}, Segment{Pos{}, Pos{10, 0}}, id, 0},
		{Segment{Pos{}, Pos{10, 0}}, Segment{Pos{10, 0}, Pos{}}, id, 0},
		{Segment{Pos{}, Pos{10, 0}}, Segment{Pos{}, Pos{20, 0}}, trans{s: 2}, 0},
		{Segment{Pos{}, Pos{10, 0}}, Segment{Pos{5, 0}, Pos{15, 0}}, trans{s: 1, t: Pos{5, 0}}, 0},
		{Segment{Pos{}, Pos{10, 0}}, Segment{Pos{0, 0}, Pos{0, 10}}, trans{s: 1, t: Pos{-5, 5}}, 50},
	}
	for _, tt := range tests {
		got := tt.U.getMapTo(tt.V)
		got1 := tt.U.scale(got).(Segment).distance(tt.V)
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got1 == tt.want1)
	}
}

func TestBezierC_distance(t *testing.T) {
	tests := []struct {
		U     Bezier
		V     Bezier
		want  trans
		want1 fl
	}{
		{
			Bezier{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, Bezier{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, id, 0,
		},
		{
			Bezier{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, Bezier{Pos{10, 0}, Pos{10, 0}, Pos{10, 0}, Pos{20, 0}}, trans{s: 1, t: Pos{10, 0}}, 0,
		},
	}
	for _, tt := range tests {
		got := tt.U.getMapTo(tt.V)
		got1 := tt.U.scale(got).(Bezier).distance(tt.V)
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got1 == tt.want1)
	}
}

func Test_distanceAtoms(t *testing.T) {
	tests := []struct {
		U    []ShapeAtom
		V    []ShapeAtom
		want fl
	}{
		{
			[]ShapeAtom{Circle{Radius: Pos{10, 10}}}, []ShapeAtom{Circle{Radius: Pos{10, 10}}}, 0,
		},
		{
			[]ShapeAtom{Circle{Radius: Pos{10, 10}}, Circle{Radius: Pos{30, 30}}},
			[]ShapeAtom{Circle{Radius: Pos{30, 30}}, Circle{Radius: Pos{10, 10}}},
			0,
		},
		{
			[]ShapeAtom{Circle{Radius: Pos{10, 10}}, Segment{End: Pos{10, 0}}},
			[]ShapeAtom{Segment{End: Pos{10, 0}}, Circle{Radius: Pos{10, 10}}},
			0,
		},
	}
	for _, tt := range tests {
		got, _ := distanceFootprints(tt.U, tt.V)
		sym, _ := distanceFootprints(tt.V, tt.U)
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got == sym)
	}
}

func TestDistanceSynthetic(t *testing.T) {
	b := ShapeFootprint{
		Segment{Start: Pos{0, 10}, End: Pos{0, 0}},     // |
		Circle{Center: Pos{3, 3}, Radius: Pos{3, 2.8}}, // o
	}
	b2 := ShapeFootprint{
		Circle{Center: Pos{4, 3}, Radius: Pos{3, 2.8}}, // o
		Segment{Start: Pos{0, 9}, End: Pos{0, 0}},      // |
	}

	d := ShapeFootprint{
		Circle{Center: Pos{3, 3}, Radius: Pos{3, 2.8}}, // o
		Segment{Start: Pos{6, 10}, End: Pos{6, 0}},     // |
	}

	db, _ := distanceFootprints(d, b)
	bd, _ := distanceFootprints(b, d)
	bb2, _ := distanceFootprints(b, b2)
	tu.Assert(t, db == bd)
	tu.Assert(t, bd > bb2)
}

func TestDistanceOo(t *testing.T) {
	o := ShapeFootprint{Circle{Radius: Pos{3, 3}}}
	O := ShapeFootprint{Circle{Radius: Pos{5, 4.5}}}

	oo, troo := distanceFootprints(o, o)
	_, trOo := distanceFootprints(o, O)
	tu.Assert(t, oo == 0)
	tu.Assert(t, troo.det() < trOo.det())
}
