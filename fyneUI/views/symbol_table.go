package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/fyneUI/whiteboard"
	sy "github.com/benoitkugler/pen2latex/symbols"
)

func showSymbolTable(onDone func(map[rune]sy.Symbol)) *fyne.Container {
	st := newSymbolTable(onDone)
	return container.NewVBox(st.title, container.NewCenter(st.wb), st.nextButton, st.doneButton)
}

type symbolTable struct {
	title      *widget.Label
	wb         *whiteboard.Whiteboard
	nextButton *widget.Button
	doneButton *widget.Button

	mapping      map[rune]sy.Symbol
	currentIndex int
}

func newSymbolTable(onDone func(map[rune]sy.Symbol)) *symbolTable {
	st := symbolTable{
		title:      widget.NewLabel(""),
		wb:         whiteboard.NewWhiteboard(),
		nextButton: widget.NewButton("Suivant", nil),
		doneButton: widget.NewButton("Terminer", nil),

		currentIndex: -1,
		mapping:      make(map[rune]sy.Symbol),
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
	rec := st.wb.Recorder.Record
	st.wb.Content = []sy.Symbol{sy.Symbol(rec)}

	// atoms := rec.Compound().SegmentToAtoms()
	// fmt.Println(len(atoms), atoms)
	// imag := renderAtoms(atoms, rec.Compound().Union().BoundingBox())
	// savePng(imag)
}

func (st *symbolTable) saveRune() {
	r := sy.RequiredRunes[st.currentIndex]
	st.mapping[r] = sy.Symbol(st.wb.Recorder.Record)
	st.wb.Recorder.Reset()
	st.wb.Content = nil
	st.wb.Refresh()
}

func (st *symbolTable) showNextRune() {
	st.currentIndex += 1
	if st.currentIndex >= len(sy.RequiredRunes) {
		st.title.SetText("Terminé.")
		st.nextButton.Disable()
		st.doneButton.Enable()
		return
	}
	st.title.SetText("Entrer le caractère " + string(sy.RequiredRunes[st.currentIndex]))
}
