package views

import (
	"image/png"
	"os"
	"testing"

	"github.com/benoitkugler/pen2latex/symbols"
	"github.com/benoitkugler/pen2latex/testutils"
)

func TestRenderAtom(t *testing.T) {
	img := renderAtoms([]symbols.ShapeAtom{
		symbols.BezierC{symbols.Pos{300, 50}, symbols.Pos{300, 300}, symbols.Pos{400, 300}, symbols.Pos{100, 300}},
		symbols.Circle{symbols.Pos{50, 50}, 30},
		symbols.Segment{Start: symbols.Pos{20, 20}, End: symbols.Pos{60, 60}},
	}, symbols.Rect{UL: symbols.Pos{}, LR: symbols.Pos{350, 350}})
	out, err := os.Create("tmp.png")
	testutils.AssertNoErr(t, err)

	err = png.Encode(out, img)
	testutils.AssertNoErr(t, err)
}
