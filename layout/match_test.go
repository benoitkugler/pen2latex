package layout

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/pen2latex/symbols"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

func rect(xLeft, xRight, yTop, yBottom float32) symbols.Rect {
	return symbols.Rect{UL: symbols.Pos{X: xLeft, Y: yTop}, LR: symbols.Pos{X: xRight, Y: yBottom}}
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

func TestInsertIndex(t *testing.T) {
	glyph := sy.Rect{UL: sy.Pos{X: 75.0, Y: 88.3}, LR: sy.Pos{X: 96.0, Y: 140.3}}
	candidates := []sy.Rect{
		{UL: sy.Pos{X: 67.0, Y: 85.5}, LR: sy.Pos{X: 114.1, Y: 133.5}},
	}
	fmt.Println(indexInsertRectBetweenArea(glyph, candidates))
}
