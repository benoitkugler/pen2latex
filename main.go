package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/unit"
	"github.com/benoitkugler/pen2latex/GUI"
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("Pen2LaTeX"),
			app.Size(unit.Dp(400), unit.Dp(600)),
		)
		err := GUI.Run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
