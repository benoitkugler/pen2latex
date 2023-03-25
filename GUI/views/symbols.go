package views

import (
	"fmt"
	"image"
	"image/color"
	"unicode/utf8"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	sh "github.com/benoitkugler/pen2latex/GUI/shared"
	"github.com/benoitkugler/pen2latex/GUI/whiteboard"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

type Store struct {
	theme *material.Theme

	store       *sy.Store
	list        widget.List
	editButtons []widget.Clickable

	runeField widget.Editor
	addButton widget.Clickable

	indexEdited int
	editor      *whiteboard.Whiteboard
	validButton widget.Clickable
	resetButton widget.Clickable

	BackButton widget.Clickable
}

func NewStore(store *sy.Store, th *material.Theme) Store {
	out := Store{store: store, theme: th}
	out.list.Axis = layout.Vertical
	out.editButtons = make([]widget.Clickable, len(store.Symbols))

	out.editor = whiteboard.NewWhiteboard(th)
	out.indexEdited = -1

	out.runeField = widget.Editor{Alignment: text.Middle, SingleLine: true, Submit: true, MaxLen: 1}

	return out
}

func (fl *Store) Layout(gtx C) D {
	// event handling

	if fl.validButton.Clicked() {
		// commit the changes
		fl.store.Symbols[fl.indexEdited].Footprint = fl.editor.Footprint()
		fl.editor.Reset()
		fl.indexEdited = -1
	}

	if fl.resetButton.Clicked() {
		fl.editor.Reset()
	}

	if fl.addButton.Clicked() {
		r, _ := utf8.DecodeRuneInString(fl.runeField.Text())
		fl.runeField.SetText("")

		fl.store.Symbols = append(fl.store.Symbols, sy.RuneFootprint{R: r})
		fl.editButtons = append(fl.editButtons, widget.Clickable{})
		fl.indexEdited = len(fl.store.Symbols) - 1
	}

	if fl.indexEdited != -1 { // editor mode
		return sh.Padding(10).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx, layout.Rigid(fl.editor.Layout))
				}),
				layout.Rigid(sh.WithPadding(10, material.Button(fl.theme, &fl.resetButton, "Effacer").Layout)),
				layout.Rigid(sh.WithPadding(10, material.Button(fl.theme, &fl.validButton, "Enregistrer le symbol").Layout)),
			)
		})
	}

	// list mode
	add := material.Button(fl.theme, &fl.addButton, "Ajouter un symbol")
	add.Background = color.NRGBA{10, 200, 10, 255}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(8, func(gtx layout.Context) layout.Dimensions {
			return material.List(fl.theme, &fl.list).Layout(gtx, len(fl.store.Symbols), func(gtx C, index int) D {
				item := fl.store.Symbols[index]
				btn := &fl.editButtons[index]
				if btn.Clicked() {
					// reset to not mix runes
					fl.editor.Reset()
					fl.indexEdited = index
				}
				return layoutFootprintCard(item, btn, gtx, fl.theme)
			})
		}),
		layout.Rigid(sh.WithPadding(10, func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					if fl.runeField.Len() == 0 {
						return add.Layout(gtx.Disabled())
					}
					return add.Layout(gtx)
				}),
				layout.Rigid(sh.WithPadding(5, material.Editor(fl.theme, &fl.runeField, "Nouveau symbol").Layout)),
			)
		})),
		layout.Rigid(material.Button(fl.theme, &fl.BackButton, "Retour").Layout),
	)
}

func layoutFootprintCard(fp sy.RuneFootprint, btn *widget.Clickable, gtx C, th *material.Theme) D {
	borderColor := color.NRGBA{0, 200, 100, 255}
	border := widget.Border{Color: borderColor, CornerRadius: 10, Width: 1}
	return sh.Padding(5).Layout(gtx, func(gtx C) D {
		return border.Layout(gtx, func(gtx C) D {
			return sh.Padding(10).Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, sh.Column(
						layout.Rigid(material.Body1(th, fmt.Sprintf("%s\n(u+%04X)", string(fp.R), fp.R)).Layout),
						layout.Rigid(layout.Spacer{Height: 20}.Layout),
						layout.Rigid(material.Button(th, btn, "Modifier").Layout),
					)),
					layout.Rigid(sh.WithPadding(20, footprint{fp.Footprint}.layout)),
				)
			})
		})
	})
}

type footprint struct {
	fp sy.Footprint
}

const footprintWidth = 100

func (b footprint) layout(gtx C) D {
	// transform the footprint to a 100x100 box
	bbox := b.fp.BoundingBox()
	maxDim := sy.Max(bbox.Width(), bbox.Height())
	scaleFactor := footprintWidth / maxDim
	tr := bbox.UL.ScaleTo(scaleFactor).ScaleTo(-1)
	trans := sy.Trans{Scale: scaleFactor, Translation: tr}

	whiteboard.DrawFootprint(gtx.Ops, b.fp, trans)

	return D{Size: image.Pt(footprintWidth, footprintWidth)}
}
