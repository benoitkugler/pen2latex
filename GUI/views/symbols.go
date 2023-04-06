package views

import (
	"fmt"
	"image"
	"image/color"
	"unicode/utf8"

	"gioui.org/layout"
	"gioui.org/op"
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

const (
	viewList uint8 = iota
	viewEdit
	viewCreate
)

type Store struct {
	theme *material.Theme

	store *sy.Store

	viewKind uint8
	list     symbolList
	editor   symbolEditor
	creator  storeCreation

	BackButton widget.Clickable
}

type symbolList struct {
	list        widget.List
	editButtons []widget.Clickable

	runeField        widget.Editor
	addButton        widget.Clickable
	resetStoreButton widget.Clickable
}

type symbolEditor struct {
	editor      *whiteboard.Whiteboard
	validButton widget.Clickable
	resetButton widget.Clickable
	backButton  widget.Clickable
	index       int
}

func NewStore(store *sy.Store, th *material.Theme) Store {
	out := Store{store: store, theme: th}

	out.list.list.Axis = layout.Vertical
	out.list.editButtons = make([]widget.Clickable, len(store.Symbols))
	out.list.runeField = widget.Editor{Alignment: text.Middle, SingleLine: true, Submit: true, MaxLen: 1}

	out.editor.editor = whiteboard.NewWhiteboard(th)

	return out
}

func (fl *Store) Layout(gtx C) D {
	// event handling
	if fl.list.resetStoreButton.Clicked() {
		// start a new store creation
		fl.creator = newStoreCreation(fl.theme)

		fl.viewKind = viewCreate
	}

	if fl.viewKind == viewCreate && fl.creator.isDone() {
		// use the new store
		*fl.store = sy.NewStore(fl.creator.symbols)
		fl.list.editButtons = make([]widget.Clickable, len(fl.store.Symbols))

		fl.viewKind = viewList
	}

	if fl.editor.validButton.Clicked() {
		// commit the changes
		fmt.Println(fl.editor.editor.Record())

		fl.store.Symbols[fl.editor.index].Footprint = fl.editor.editor.Footprint()
		fl.editor.editor.Reset()

		fl.viewKind = viewList
	}

	if fl.editor.resetButton.Clicked() {
		fl.editor.editor.Reset()
	}
	if fl.editor.backButton.Clicked() { // cancel editing
		fl.viewKind = viewList
	}

	if fl.list.addButton.Clicked() {
		r, _ := utf8.DecodeRuneInString(fl.list.runeField.Text())
		fl.list.runeField.SetText("")

		fl.store.Symbols = append(fl.store.Symbols, sy.RuneFootprint{R: r})
		fl.list.editButtons = append(fl.list.editButtons, widget.Clickable{})
		fl.editor.index = len(fl.store.Symbols) - 1

		fl.viewKind = viewEdit
	}

	switch fl.viewKind {
	case viewList:
		return fl.layoutList(gtx)
	case viewEdit:
		return fl.layoutEdit(gtx)
	case viewCreate:
		return fl.layoutCreate(gtx)
	default:
		panic("exhaustive switch")
	}
}

// list mode
func (fl *Store) layoutList(gtx C) D {
	add := sh.Button(fl.theme, &fl.list.addButton, "Ajouter un symbol", sh.PositiveAction)

	reset := sh.Button(fl.theme, &fl.list.resetStoreButton, "Ré-initialiser", sh.NegativeAction)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(8, func(gtx layout.Context) layout.Dimensions {
			return material.List(fl.theme, &fl.list.list).Layout(gtx, len(fl.store.Symbols), func(gtx C, index int) D {
				item := fl.store.Symbols[index]
				btn := &fl.list.editButtons[index]
				if btn.Clicked() {
					// reset to not mix runes
					fl.editor.editor.Reset()
					fl.editor.index = index

					fl.viewKind = viewEdit
				}
				return layoutFootprintCard(item, btn, gtx, fl.theme)
			})
		}),
		layout.Rigid(sh.WithPadding(5, func(gtx C) D {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					if fl.list.runeField.Len() == 0 {
						return add.Layout(gtx.Disabled())
					}
					return add.Layout(gtx)
				}),
				layout.Rigid(sh.WithPadding(5, material.Editor(fl.theme, &fl.list.runeField, "Nouveau symbol").Layout)),
			)
		})),
		layout.Rigid(sh.WithPadding(5, reset.Layout)),
		layout.Rigid(material.Button(fl.theme, &fl.BackButton, "Retour").Layout),
	)
}

func (fl *Store) layoutEdit(gtx C) D {
	return sh.Padding(10).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx, layout.Rigid(fl.editor.editor.Layout))
			}),
			layout.Rigid(sh.WithPadding(10, sh.Button(fl.theme, &fl.editor.resetButton, "Effacer", sh.NegativeAction).Layout)),
			layout.Rigid(sh.WithPadding(10, sh.Button(fl.theme, &fl.editor.validButton, "Enregistrer le symbol", sh.PositiveAction).Layout)),
			layout.Rigid(sh.WithPadding(10, material.Button(fl.theme, &fl.editor.backButton, "Retour").Layout)),
		)
	})
}

func (fl *Store) layoutCreate(gtx C) D { return fl.creator.layout(gtx) }

type storeCreation struct {
	theme *material.Theme

	toRegister    []rune             // the symbols the user have to provide
	currentSymbol int                // index  in [toRegister]
	symbols       map[rune]sy.Symbol // the symbol drawn so far

	editor      *whiteboard.Whiteboard
	validButton widget.Clickable // go to next
	resetButton widget.Clickable
}

var requiredRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_×+-=()[]∈Σℝπ")

// var requiredRunes = []rune("abcdefxySoit()123_∈Σℝ")

// var requiredRunes = []rune("a")

func newStoreCreation(theme *material.Theme) storeCreation {
	return storeCreation{
		toRegister: requiredRunes,
		theme:      theme, symbols: make(map[rune]sy.Symbol),
		editor: whiteboard.NewWhiteboard(theme),
	}
}

func (sc *storeCreation) isDone() bool {
	return sc.currentSymbol >= len(sc.toRegister)
}

func (sc *storeCreation) layout(gtx C) D {
	// event handling

	if sc.validButton.Clicked() {
		r := sc.toRegister[sc.currentSymbol]
		sc.symbols[r] = sy.Symbol(sc.editor.Record())

		sc.currentSymbol++
		sc.editor.Reset()
	}
	if sc.resetButton.Clicked() {
		sc.editor.Reset()
	}

	if sc.isDone() {
		op.InvalidateOp{}.Add(gtx.Ops)
		return layout.Dimensions{}
	}

	currentRune := sc.toRegister[sc.currentSymbol]

	return sh.Padding(10).Layout(gtx, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.H5(sc.theme, "Symbol : "+string(currentRune)).Layout),
			layout.Rigid(func(gtx C) D {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx, layout.Rigid(sc.editor.Layout))
			}),
			layout.Rigid(sh.WithPadding(10, material.Button(sc.theme, &sc.resetButton, "Effacer").Layout)),
			layout.Rigid(sh.WithPadding(10, material.Button(sc.theme, &sc.validButton, "Continuer").Layout)),
		)
	})
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
