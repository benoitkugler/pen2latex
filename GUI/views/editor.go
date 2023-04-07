package views

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	sh "github.com/benoitkugler/pen2latex/GUI/shared"
	la "github.com/benoitkugler/pen2latex/layout"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

// Editor shows the user input and the layout boxes.
type Editor struct {
	theme *material.Theme

	store *sy.Store

	// BackButton is clicked to go back to the home view.
	BackButton widget.Clickable

	// the tree storing the current user input
	line *la.Line

	// the current context
	context sy.Context
}

const (
	width  = 600
	height = 200
)

func NewEditor(store *sy.Store, theme *material.Theme) *Editor {
	return &Editor{theme: theme, store: store, line: la.NewLine(sy.Rect{sy.Pos{}, sy.Pos{width, height}})}
}

func (ed *Editor) Layout(gtx C) D {
	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEvenly}.Layout(gtx,
		layout.Rigid(ed.layoutLine),
		layout.Rigid(material.Body1(ed.theme, fmt.Sprintf("Expression : %s", ed.line.LaTeX())).Layout),
		// layout.Rigid(sh.WithPadding(10, sh.Button(ed.theme, &ed.resetButton, "Effacer", sh.NegativeAction).Layout)),
		layout.Rigid(sh.WithPadding(10, material.Button(ed.theme, &ed.BackButton, "Retour").Layout)),
	)
}

func rectToRect(r sy.Rect) clip.Rect {
	return clip.Rect{Min: image.Pt(int(r.UL.X), int(r.UL.Y)), Max: image.Pt(int(r.LR.X), int(r.LR.Y))}
}

func (ed *Editor) layoutLine(gtx C) D {
	size := image.Pt(width, height)

	// Declare the tag.
	st := clip.Rect{Max: size}.Push(gtx.Ops)
	pointer.InputOp{
		Tag:   ed,
		Types: pointer.Press | pointer.Release | pointer.Drag | pointer.Enter | pointer.Move | pointer.Leave,
	}.Add(gtx.Ops)
	defer st.Pop()

	for _, ev := range gtx.Events(ed) {
		if ev, ok := ev.(pointer.Event); ok {
			switch ev.Type {
			case pointer.Move:
				ed.context = ed.line.FindContext(sy.Pos{X: ev.Position.X, Y: ev.Position.Y})
			case pointer.Leave:
				ed.context = sy.Context{}
			}
		}
	}

	// background
	paint.FillShape(gtx.Ops, color.NRGBA{0xE0, 0xF2, 0xF1, 0xFF}, clip.Rect{Max: size}.Op())

	// context
	if !ed.context.Box.IsEmpty() {
		box := rectToRect(ed.context.Box)
		paint.FillShape(gtx.Ops, color.NRGBA{0xE0, 0xF8, 0xA1, 0xFF}, box.Op())
		box.Min.Y = int(ed.context.Baseline)
		box.Max.Y = int(ed.context.Baseline) + 1
		paint.FillShape(gtx.Ops, color.NRGBA{10, 10, 0, 0xFF}, box.Op())
	}

	return D{Size: size}
}
