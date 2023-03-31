package views

import (
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/benoitkugler/pen2latex/GUI/shared"
	"github.com/benoitkugler/pen2latex/GUI/whiteboard"
	"github.com/benoitkugler/pen2latex/symbols"
)

type Editor struct {
	theme *material.Theme

	store *symbols.Store

	wb      whiteboard.Whiteboard
	matched rune

	resetButton widget.Clickable

	BackButton widget.Clickable
}

func NewEditor(store *symbols.Store, theme *material.Theme) *Editor {
	return &Editor{theme: theme, store: store}
}

func (ed *Editor) Layout(gtx C) D {
	// event handling
	if ed.resetButton.Clicked() {
		ed.wb.Reset()
		ed.matched = 0
	}

	if ok := ed.wb.HasNewShape(); ok {
		// udapte the recognized rune
		rec := ed.wb.Record()
		// symbol := rec.Identify(ed.store)
		// fmt.Println("Matching", len(symbol), "strokes")
		r, onlyLastUsed := rec.Identify(ed.store)
		ed.matched = r

		fmt.Println(rec)

		// drop the old strokes when we are sure they
		// are not part of a compound symbol
		if onlyLastUsed {
			ed.wb.DropButLast()
		}
	}

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEvenly}.Layout(gtx,
		layout.Rigid(shared.Flex(
			layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceEvenly, Alignment: layout.Middle},
			layout.Rigid(ed.wb.Layout),
			layout.Rigid(material.Body1(ed.theme, fmt.Sprintf("Caractère reconnu : %s", string(ed.matched))).Layout),
		)),
		layout.Rigid(shared.WithPadding(10, material.Button(ed.theme, &ed.resetButton, "Effacer").Layout)),
		layout.Rigid(shared.WithPadding(10, material.Button(ed.theme, &ed.BackButton, "Retour").Layout)),
	)
}
