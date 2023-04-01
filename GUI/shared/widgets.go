package shared

import (
	"image/color"

	"gioui.org/widget"
	"gioui.org/widget/material"
)

func Button(th *material.Theme, state *widget.Clickable, text string, color color.NRGBA) material.ButtonStyle {
	out := material.Button(th, state, text)
	out.Background = color
	return out
}
