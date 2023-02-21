package gioui

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/benoitkugler/pen2latex/giouiUI/whiteboard"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("Pen2LaTeX"),
			app.Size(unit.Dp(400), unit.Dp(600)),
		)
		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var m menu
	var ops op.Ops
	whiteboard := &whiteboard.Whiteboard{}
	var inEditor bool
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			if m.editorBtn.Clicked() {
				inEditor = true
			}

			if inEditor {
				whiteboard.Layout(layout.NewContext(&ops, e))
				e.Frame(&ops)
			} else {
				m.draw(e, &ops, th)
			}
		}
	}
}

func withPadding(w layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(25),
			Bottom: unit.Dp(25),
			Right:  unit.Dp(35),
			Left:   unit.Dp(35),
		}.Layout(gtx, w)
	}
}

type menu struct {
	tableBtn  widget.Clickable
	editorBtn widget.Clickable
}

func (menu *menu) draw(e system.FrameEvent, ops *op.Ops, th *material.Theme) {
	gtx := layout.NewContext(ops, e)

	title := material.H1(th, "Hello, Gio")
	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}
	title.Color = maroon
	title.Alignment = text.Middle
	title.Layout(gtx)

	layout.Flex{
		// Vertical alignment, from top to bottom
		Axis: layout.Vertical,
		// Empty space is left at the start, i.e. at the top
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(withPadding(func(gtx layout.Context) layout.Dimensions {
			return material.Button(th, &menu.tableBtn, "Créer la table des caractères...").Layout(gtx)
		}),
		),
		layout.Rigid(withPadding(func(gtx C) D {
			return material.Button(th, &menu.editorBtn, "Rédiger...").Layout(gtx)
		})),
	)

	e.Frame(gtx.Ops)
}
