package symbols

import (
	"reflect"
	"testing"
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
