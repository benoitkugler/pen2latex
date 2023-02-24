package views

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/fyneUI/whiteboard"
	"github.com/benoitkugler/pen2latex/layout"
	"github.com/benoitkugler/pen2latex/symbols"
)

func showEditor(db *symbols.SymbolStore) *fyne.Container {
	// ed := newEditor(db)

	data := canvas.NewImageFromImage(image.NewGray(image.Rect(0, -180, 100, 180)))
	data.FillMode = canvas.ImageFillOriginal

	shapeImg := canvas.NewImageFromImage(image.NewGray(image.Rect(0, -180, 100, 180)))
	shapeImg.FillMode = canvas.ImageFillOriginal
	rec := whiteboard.NewRecorder()
	rec.OnEndShape = func() {
		img := rec.Recorder.Current().Shape().AngularGraph()
		data.Image = img
		data.Refresh()

		si := rec.Recorder.Current().Shape().AngleClustersGraph()
		shapeImg.Image = si
		shapeImg.Refresh()

		fmt.Println(rec.Recorder.Current().Shape().Segment())
	}
	return container.NewVBox(rec, shapeImg, data)
}

type editor struct {
	wb          *whiteboard.Whiteboard
	recognized  *widget.Label
	resetButton *widget.Button

	db *symbols.SymbolStore

	line layout.Line
}

func newEditor(db *symbols.SymbolStore) *editor {
	ed := &editor{
		wb:          whiteboard.NewWhiteboard(),
		recognized:  widget.NewLabel("Dessiner un caractère..."),
		resetButton: widget.NewButton("Effacer", nil),
		db:          db,
	}
	ed.wb.OnEndShape = ed.tryMatchShape
	ed.wb.OnCursorMove = ed.showScope
	ed.resetButton.OnTapped = ed.clear
	return ed
}

func (ed *editor) tryMatchShape() {
	rec := ed.wb.Recorder.Current()
	if len(rec.Compound()) == 0 {
		return
	}
	ed.line.Insert(rec, ed.db)
	ed.wb.Content = ed.line.Symbols()
	ed.wb.Scopes = ed.line.Scopes()
	ed.recognized.SetText(ed.line.LaTeX())
}

func (ed *editor) showScope(pos symbols.Pos) {
	glyph := symbols.Rect{
		UL: symbols.Pos{X: pos.X - 1, Y: pos.Y - 1},
		LR: symbols.Pos{X: pos.X + 1, Y: pos.Y + 1},
	}
	_, scope, _ := ed.line.FindNode(glyph)
	if scope.IsEmpty() { // root
		scope = ed.wb.RootScope()
	}
	ed.wb.HighlightedScope = scope
}

func (ed *editor) clear() {
	ed.wb.Recorder.Reset()
	ed.line = layout.Line{}
	ed.wb.Content = nil
	ed.wb.Scopes = nil
	ed.wb.Refresh()
	ed.recognized.SetText("Dessiner un caractère...")
}
