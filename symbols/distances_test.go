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
		U     BezierC
		V     BezierC
		want  trans
		want1 fl
	}{
		{
			BezierC{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, BezierC{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, id, 0,
		},
		{
			BezierC{Pos{}, Pos{}, Pos{}, Pos{10, 0}}, BezierC{Pos{10, 0}, Pos{10, 0}, Pos{10, 0}, Pos{20, 0}}, trans{s: 1, t: Pos{10, 0}}, 0,
		},
	}
	for _, tt := range tests {
		got := tt.U.getMapTo(tt.V)
		got1 := tt.U.scale(got).(BezierC).distance(tt.V)
		tu.Assert(t, got == tt.want)
		tu.Assert(t, got1 == tt.want1)
	}
}

func Test_bestShapeDistance(t *testing.T) {
	tests := []struct {
		U    []ShapeAtom
		V    []ShapeAtom
		want fl
	}{
		{
			[]ShapeAtom{Circle{Radius: 10}}, []ShapeAtom{Circle{Radius: 10}}, 0,
		},
		{
			[]ShapeAtom{Circle{Radius: 10}, Circle{Radius: 30}},
			[]ShapeAtom{Circle{Radius: 30}, Circle{Radius: 10}},
			0,
		},
		{
			[]ShapeAtom{Circle{Radius: 10}, Segment{End: Pos{10, 0}}},
			[]ShapeAtom{Segment{End: Pos{10, 0}}, Circle{Radius: 10}},
			0,
		},
	}
	for _, tt := range tests {
		got := bestShapeDistance(tt.U, tt.V)
		tu.Assert(t, got == tt.want)
	}
}
