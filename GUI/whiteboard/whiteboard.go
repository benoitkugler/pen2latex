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
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/benoitkugler/pen2latex/GUI/shared"
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

	reset widget.Clickable

	OnValid widget.Clickable
}

func NewWhiteboard(theme *material.Theme) *Whiteboard {
	return &Whiteboard{theme: theme}
}

func (b *Whiteboard) Layout(gtx C) D {
	if b.reset.Clicked() {
		b.Reset()
	}

	return shared.Padding(10).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx, layout.Rigid(func(gtx C) D { return b.layoutDrawArea(gtx) }))
			}),
			layout.Rigid(shared.WithPadding(10, material.Button(b.theme, &b.reset, "Effacer").Layout)),
			layout.Rigid(shared.WithPadding(10, material.Button(b.theme, &b.OnValid, "Valider").Layout)),
		)
	})
}

// Footprint is the current footprint drawn
func (b *Whiteboard) Footprint() sy.Footprint { return b.footprint }

func (b *Whiteboard) Reset() {
	b.recorder.Reset()
	b.footprint = nil
}

func (b *Whiteboard) layoutDrawArea(gtx C) D {
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
