package shared

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Padding builds a uniform padding layout.
func Padding(padding int) layout.Inset {
	return layout.Inset{
		Top:    unit.Dp(padding),
		Bottom: unit.Dp(padding),
		Right:  unit.Dp(padding),
		Left:   unit.Dp(padding),
	}
}

// WithPadding add a uniform padding around w
func WithPadding(pad int, w layout.Widget) layout.Widget {
	return func(gtx C) D { return Padding(pad).Layout(gtx, w) }
}

func Flex(flex layout.Flex, children ...layout.FlexChild) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return flex.Layout(gtx, children...)
	}
}

func Column(children ...layout.FlexChild) layout.Widget {
	return Flex(layout.Flex{Axis: layout.Vertical}, children...)
}

func Row(children ...layout.FlexChild) layout.Widget {
	return Flex(layout.Flex{Axis: layout.Horizontal}, children...)
}
