package views

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/benoitkugler/pen2latex/GUI/whiteboard"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func padding(padding int) layout.Inset {
	return layout.Inset{
		Top:    unit.Dp(padding),
		Bottom: unit.Dp(padding),
		Right:  unit.Dp(padding),
		Left:   unit.Dp(padding),
	}
}

func withPadding(w layout.Widget, pad int) layout.Widget {
	return func(gtx C) D { return padding(pad).Layout(gtx, w) }
}

type Store struct {
	theme *material.Theme
	store sy.Store
	list  widget.List
}

func NewStore(store sy.Store, th *material.Theme) Store {
	out := Store{store: store, theme: th}
	out.list.Axis = layout.Vertical
	return out
}

func (fl *Store) Layout(gtx C) D {
	return material.List(fl.theme, &fl.list).Layout(gtx, len(fl.store.Symbols), func(gtx C, index int) D {
		item := fl.store.Symbols[index]
		return footprintCard(item).layout(gtx, fl.theme)
	})
}

type footprintCard sy.RuneFootprint

func (fc footprintCard) layout(gtx C, th *material.Theme) D {
	return padding(10).Layout(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(3, material.H4(th, fmt.Sprintf("Rune : %s", string(fc.R))).Layout),
			layout.Flexed(1, withPadding(footprint{fc.Footprint}.layout, 10)))
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

	return layout.Dimensions{Size: image.Pt(footprintWidth, footprintWidth)}
}
