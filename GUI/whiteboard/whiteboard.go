package whiteboard

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	la "github.com/benoitkugler/pen2latex/layout"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

type Whiteboard struct {
	resolvedFootprint sy.Footprint

	recorder la.Recorder
}

func (b *Whiteboard) Layout(gtx layout.Context) layout.Dimensions {
	// Declare the tag.
	st := clip.Rect{Max: image.Pt(100, 100)}.Push(gtx.Ops)
	pointer.InputOp{
		Tag:   &b.recorder,
		Types: pointer.Press | pointer.Release | pointer.Drag | pointer.Enter | pointer.Move,
	}.Add(gtx.Ops)
	defer st.Pop()

	for _, ev := range gtx.Events(&b.recorder) {
		if x, ok := ev.(pointer.Event); ok {
			switch x.Type {
			case pointer.Press:
				b.recorder.StartShape()
			case pointer.Release:
				b.recorder.EndShape()
				b.resolvedFootprint = sy.Symbol(b.recorder.Record).Footprint()
			case pointer.Drag:
				b.recorder.AddToShape(sy.Pos{X: x.Position.X, Y: x.Position.Y})
			}
		}
	}

	drawSquare(gtx.Ops)
	DrawFootprint(gtx.Ops, b.resolvedFootprint, sy.Id)

	return layout.Dimensions{Size: image.Pt(100, 100)}
}

func drawSquare(ops *op.Ops) {
	paint.FillShape(ops, color.NRGBA{255, 255, 12, 255}, clip.Rect{Max: image.Pt(100, 100)}.Op())
}

// DrawFootprint draw the given footprint, with [trans] applied
func DrawFootprint(ops *op.Ops, footprint sy.Footprint, trans sy.Trans) {
	color := color.NRGBA{255, 12, 12, 255}

	for _, shape := range footprint {
		var path clip.Path
		path.Begin(ops)
		for _, curve := range shape.Curves {
			curve = curve.Scale(trans)
			path.MoveTo(f32.Pt(curve.P0.X, curve.P0.Y))
			path.CubeTo(f32.Pt(curve.P1.X, curve.P1.Y), f32.Pt(curve.P2.X, curve.P2.Y), f32.Pt(curve.P3.X, curve.P3.Y))
		}
		spec := path.End()

		paint.FillShape(ops, color, clip.Stroke{Path: spec, Width: 1.2}.Op())
	}
}
