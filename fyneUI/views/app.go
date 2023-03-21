package views

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/symbols"
)

func Home() *fyne.Container {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err) // TODO:
	}
	storePath := filepath.Join(homeDir, "pen2latex.store.json")

	main := container.NewVBox()

	database, _ := symbols.NewStoreFromDisk(storePath) // TODO:
	var showButtons func()

	showButtons = func() {
		main.RemoveAll()
		main.Add(
			widget.NewButton("Créer la table des caractères...", func() {
				main.RemoveAll()
				main.Add(showSymbolTable(func(m map[rune]symbols.Symbol) {
					database = symbols.NewStore(m)
					err = database.Serialize(storePath)
					if err != nil {
						panic(err) // TODO:
					}
					showButtons()
				}))
				main.Refresh()
			}),
		)
		main.Add(
			widget.NewButton("Rédiger...", func() {
				main.RemoveAll()
				main.Add(showEditor(database))
				main.Refresh()
			}),
		)
		main.Refresh()
	}

	showButtons()

	return container.NewVBox(
		widget.NewLabel("Pen to LaTeX - Menu principal"),
		main,
	)
}
