package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/fyneUI/whiteboard"
	"github.com/benoitkugler/pen2latex/symbols"
)

func showSymbolTable(onDone func(map[rune]symbols.Symbol)) *fyne.Container {
	st := newSymbolTable(onDone)
	return container.NewVBox(st.title, container.NewCenter(st.wb), st.nextButton, st.doneButton)
}

type symbolTable struct {
	title      *widget.Label
	wb         *whiteboard.Whiteboard
	nextButton *widget.Button
	doneButton *widget.Button

	mapping      map[rune]symbols.Symbol
	currentIndex int
}

func newSymbolTable(onDone func(map[rune]symbols.Symbol)) *symbolTable {
	st := symbolTable{
		title:      widget.NewLabel(""),
		wb:         whiteboard.NewWhiteboard(),
		nextButton: widget.NewButton("Suivant", nil),
		doneButton: widget.NewButton("Terminer", nil),

		currentIndex: -1,
		mapping:      make(map[rune]symbols.Symbol),
	}
	st.nextButton.OnTapped = func() {
		st.saveRune()
		st.showNextRune()
	}
	st.doneButton.OnTapped = func() { onDone(st.mapping) }
	st.wb.OnEndShape = st.drawShape

	// start
	st.doneButton.Disable()
	st.showNextRune()
	return &st
}

func (st *symbolTable) drawShape() {
	rec := st.wb.Recorder.Current()
	st.wb.Content = []symbols.Symbol{rec.Compound()}
}

func (st *symbolTable) saveRune() {
	r := symbols.RequiredRunes[st.currentIndex]
	st.mapping[r] = st.wb.Recorder.Current().Compound()
	st.wb.Recorder.Reset()
	st.wb.Content = nil
	st.wb.Refresh()
}

func (st *symbolTable) showNextRune() {
	st.currentIndex += 1
	if st.currentIndex >= len(symbols.RequiredRunes) {
		st.title.SetText("Terminé.")
		st.nextButton.Disable()
		st.doneButton.Enable()
		return
	}
	st.title.SetText("Entrer le caractère " + string(symbols.RequiredRunes[st.currentIndex]))
}
