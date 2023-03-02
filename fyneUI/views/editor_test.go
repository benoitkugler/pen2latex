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
		symbols.Bezier{symbols.Pos{300, 50}, symbols.Pos{300, 300}, symbols.Pos{400, 300}, symbols.Pos{100, 300}},
		symbols.Circle{symbols.Pos{50, 50}, symbols.Pos{30, 30}},
		symbols.Segment{Start: symbols.Pos{20, 20}, End: symbols.Pos{60, 60}},
	}, symbols.Rect{UL: symbols.Pos{}, LR: symbols.Pos{350, 350}})
	out, err := os.Create("tmp.png")
	testutils.AssertNoErr(t, err)

	err = png.Encode(out, img)
	testutils.AssertNoErr(t, err)
}

func TestDebug(t *testing.T) {
	a := symbols.Symbol{{
		{X: 24.00, Y: 18.30}, {X: 23.00, Y: 17.30}, {X: 22.00, Y: 17.30}, {X: 21.00, Y: 17.30}, {X: 20.00, Y: 17.30}, {X: 19.00, Y: 17.30}, {X: 18.00, Y: 17.30}, {X: 18.00, Y: 18.30}, {X: 17.00, Y: 18.30}, {X: 17.00, Y: 20.30}, {X: 16.00, Y: 21.30}, {X: 16.00, Y: 22.30}, {X: 16.00, Y: 24.30}, {X: 16.00, Y: 25.30}, {X: 17.00, Y: 27.30}, {X: 17.00, Y: 28.30}, {X: 18.00, Y: 29.30}, {X: 19.00, Y: 29.30}, {X: 20.00, Y: 29.30}, {X: 21.00, Y: 29.30}, {X: 22.00, Y: 29.30}, {X: 23.00, Y: 28.30}, {X: 23.00, Y: 26.30}, {X: 24.00, Y: 25.30}, {X: 24.00, Y: 23.30}, {X: 24.00, Y: 21.30}, {X: 24.00, Y: 19.30}, {X: 24.00, Y: 17.30}, {X: 24.00, Y: 16.30}, {X: 24.00, Y: 15.30}, {X: 23.00, Y: 15.30}, {X: 23.00, Y: 16.30}, {X: 23.00, Y: 17.30}, {X: 23.00, Y: 19.30}, {X: 23.00, Y: 21.30}, {X: 24.00, Y: 23.30}, {X: 25.00, Y: 24.30}, {X: 26.00, Y: 26.30}, {X: 27.00, Y: 27.30}, {X: 28.00, Y: 28.30}, {X: 29.00, Y: 28.30}, {X: 30.00, Y: 29.30}, {X: 31.00, Y: 29.30}, {X: 32.00, Y: 29.30},
	}}
	img := renderAtoms(a.SegmentToAtoms(), a.Union().BoundingBox())

	img = renderAtoms([]symbols.ShapeAtom{
		symbols.Bezier{symbols.Pos{X: 24.0, Y: 18.3}, symbols.Pos{X: 7.6, Y: 1.9}, symbols.Pos{X: 24.0, Y: 56.0}, symbols.Pos{X: 24.0, Y: 16.3}},
		symbols.Circle{Center: symbols.Pos{X: 20.6, Y: 23.0}, Radius: symbols.Pos{5.65608, 5}},
	}, symbols.Rect{symbols.Pos{}, symbols.Pos{40, 40}})
	out, err := os.Create("tmp.png")
	testutils.AssertNoErr(t, err)

	err = png.Encode(out, img)
	testutils.AssertNoErr(t, err)
}
