package layout

import (
	"testing"

	"github.com/benoitkugler/pen2latex/symbols"
)

func rect(xLeft, xRight, yTop, yBottom float32) symbols.Rect {
	return symbols.Rect{UL: symbols.Pos{xLeft, yTop}, LR: symbols.Pos{xRight, yBottom}}
}

func Test_isRectInAreas(t *testing.T) {
	tests := []struct {
		glyph      symbols.Rect
		candidates []symbols.Rect
		want       int
	}{
		{
			rect(49, 58, 41, 49),
			[]symbols.Rect{
				rect(31, 62, -14, 56),
			},
			0,
		},
		{
			rect(49, 58, 41, 49),
			[]symbols.Rect{
				rect(25, 30, -14, 56),
				rect(31, 62, -14, 56),
			},
			1,
		},
	}
	for _, tt := range tests {
		if got := isRectInAreas(tt.glyph, tt.candidates); got != tt.want {
			t.Errorf("isRectInAreas() = %v, want %v", got, tt.want)
		}
	}
}
