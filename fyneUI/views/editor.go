package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/fyneUI/whiteboard"
	"github.com/benoitkugler/pen2latex/layout"
	"github.com/benoitkugler/pen2latex/symbols"
)

func showEditor(db *symbols.SymbolStore) *fyne.Container {
	ed := newEditor(db)
	return container.NewVBox(ed.wb, ed.recognized, ed.resetButton)
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
}

func (ed *editor) showScope(pos symbols.Pos) {
	glyph := symbols.Rect{
		UL: symbols.Pos{X: pos.X - 1, Y: pos.Y - 1},
		LR: symbols.Pos{X: pos.X + 1, Y: pos.Y + 1},
	}
	_, scope, _ := ed.line.FindNode(glyph)
	ed.wb.HighlightedScope = scope
}

func (ed *editor) clear() {
	ed.wb.Recorder.Reset()
	ed.line = layout.Line{}
	ed.wb.Content = nil
	ed.wb.Scopes = nil
	ed.wb.Refresh()
}
