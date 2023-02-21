package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/benoitkugler/pen2latex/fyneUI/views"
)

func main() {
	a := app.New()
	w := a.NewWindow("Pen to LaTeX")

	w.SetContent(views.Home())
	w.ShowAndRun()
}
