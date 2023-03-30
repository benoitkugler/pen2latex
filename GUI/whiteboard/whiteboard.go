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
	"gioui.org/widget/material"
	la "github.com/benoitkugler/pen2latex/layout"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type Whiteboard struct {
	theme *material.Theme

	// cached value
	footprint sy.Footprint

	recorder la.Recorder
	newShape bool
}

func NewWhiteboard(theme *material.Theme) *Whiteboard { return &Whiteboard{theme: theme} }

// Footprint is the current footprint drawn
func (b *Whiteboard) Footprint() sy.Footprint { return b.footprint }
func (b *Whiteboard) Record() la.Record       { return b.recorder.Record }

// Reset remove any drawings or recordings
func (b *Whiteboard) Reset() {
	b.recorder.Reset()
	b.footprint = sy.Footprint{}
}

func (b *Whiteboard) DropButLast() {
	b.recorder.DropButLast()
	b.footprint = sy.Symbol(b.recorder.Record).Footprint()
}

// HasNewShape is the event for a new shape.
func (b *Whiteboard) HasNewShape() bool {
	if b.newShape {
		b.newShape = false
		return true
	}
	return false
}

func (b *Whiteboard) Layout(gtx C) D {
	size := image.Pt(120, 140)

	// Declare the tag.
	st := clip.Rect{Max: size}.Push(gtx.Ops)
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
				b.footprint = sy.Symbol(b.recorder.Record).Footprint()
				b.newShape = true
			case pointer.Drag:
				b.recorder.AddToShape(sy.Pos{X: x.Position.X, Y: x.Position.Y})
			}
		}
	}

	// background
	paint.FillShape(gtx.Ops, color.NRGBA{0xE0, 0xF2, 0xF1, 0xFF}, clip.Rect{Max: size}.Op())
	// drawn symbol
	DrawFootprint(gtx.Ops, b.footprint, sy.Id)

	return layout.Dimensions{Size: size}
}

// DrawFootprint draw the given footprint, with [trans] applied
func DrawFootprint(ops *op.Ops, footprint sy.Footprint, trans sy.Trans) {
	color := color.NRGBA{255, 12, 12, 255}

	for _, shape := range footprint.Strokes {
		if point, ok := shape.IsPoint(); ok {
			point = trans.Apply(point)
			point = point.Sub(sy.Pos{X: 0.5, Y: 0.5})
			min := image.Point{int(point.X), int(point.Y)}
			max := min.Add(image.Point{1, 1})

			spec := clip.Ellipse{Min: min, Max: max}.Path(ops)
			paint.FillShape(ops, color, clip.Stroke{Path: spec, Width: 1.2}.Op())
		} else {
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
}
