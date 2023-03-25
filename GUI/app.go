package GUI

import (
	"image/color"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/benoitkugler/pen2latex/GUI/views"
	"github.com/benoitkugler/pen2latex/GUI/whiteboard"
	"github.com/benoitkugler/pen2latex/symbols"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func loadStore() symbols.Store {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err) // TODO:
	}
	storePath := filepath.Join(homeDir, "pen2latex.store.json")
	database, err := symbols.NewStoreFromDisk(storePath) // TODO:
	if err != nil {
		log.Println(err)
		database = symbols.NewStore(nil)
	}
	return database
}

func saveStore(st symbols.Store) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err) // TODO:
	}
	storePath := filepath.Join(homeDir, "pen2latex.store.json")
	err = st.Serialize(storePath)
	if err != nil {
		panic(err) // TODO:
	}
	log.Println("Store saved in", storePath)
}

// Run starts the application
func Run(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	store := loadStore()

	var homeMenu menu

	whiteboard := whiteboard.NewWhiteboard(th)
	symbols := views.NewStore(&store, th)

	view := viewHome

	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			saveStore(store)
			return e.Err
		case system.FrameEvent:
			if homeMenu.editorBtn.Clicked() {
				view = viewEditor
			} else if homeMenu.tableBtn.Clicked() {
				view = viewSymbols
			} else if whiteboard.OnValid.Clicked() {
				view = viewHome
			}

			gtx := layout.NewContext(&ops, e)

			switch view {
			case viewHome:
				homeMenu.layout(gtx, th)
			case viewEditor:
				whiteboard.Layout(gtx)
			case viewSymbols:
				symbols.Layout(gtx)
			}

			e.Frame(gtx.Ops)
		}
	}
}

// views id
const (
	viewHome = iota
	viewSymbols
	viewEditor
)

func withPadding(w layout.Widget) layout.Widget {
	return func(gtx C) D {
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

func (menu *menu) layout(gtx C, th *material.Theme) {
	title := material.H2(th, "Bienvenue dans Pen2LaTeX")
	title.Color = color.NRGBA{R: 10, G: 50, B: 150, A: 255}
	title.Alignment = text.Middle
	title.Layout(gtx)

	layout.Flex{
		// Vertical alignment, from top to bottom
		Axis: layout.Vertical,
		// Empty space is left at the start, i.e. at the top
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(withPadding(func(gtx C) D {
			return material.Button(th, &menu.tableBtn, "Editer la table des caractères...").Layout(gtx)
		}),
		),
		layout.Rigid(withPadding(func(gtx C) D {
			return material.Button(th, &menu.editorBtn, "Rédiger...").Layout(gtx)
		})),
	)
}
