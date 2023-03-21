package whiteboard

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	sym "github.com/benoitkugler/pen2latex/symbols"
	"golang.org/x/image/colornames"
)

var (
	_ desktop.Mouseable = (*Whiteboard)(nil)
	_ desktop.Hoverable = (*Whiteboard)(nil)

	_ fyne.WidgetRenderer = (*whiteboardRenderer)(nil)
)

type whiteboardRenderer struct {
	widget *Whiteboard

	rec             *canvas.Rectangle
	baseline        *canvas.Line
	shapeLines      []fyne.CanvasObject
	scopesRects     []*canvas.Rectangle
	higlightedScope *canvas.Rectangle
}

func newWhiteboardRendered(w *Whiteboard) *whiteboardRenderer {
	return &whiteboardRenderer{
		widget:          w,
		rec:             canvas.NewRectangle(colornames.Lightgray),
		baseline:        canvas.NewLine(color.Black),
		higlightedScope: canvas.NewRectangle(color.NRGBA{200, 10, 10, 100}),
	}
}

func (w *whiteboardRenderer) Destroy() {}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
func (w *whiteboardRenderer) Layout(size fyne.Size) {
	w.rec.Resize(size)
	w.baseline.Position1 = fyne.NewPos(0, size.Height*sym.EMBaselineRatio)
	w.baseline.Position2 = fyne.NewPos(size.Width, size.Height*sym.EMBaselineRatio)
}

func (w *whiteboardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(2*sym.EMWidth, 3*sym.EMHeight)
}

// Objects returns all objects that should be drawn.
func (w *whiteboardRenderer) Objects() []fyne.CanvasObject {
	out := make([]fyne.CanvasObject, 0, 2+len(w.shapeLines)+len(w.scopesRects))
	out = append(out, w.rec, w.baseline, w.higlightedScope)
	out = append(out, w.shapeLines...)
	for _, r := range w.scopesRects {
		out = append(out, r)
	}
	return out
}

func rectToFyne(rect sym.Rect) (fyne.Position, fyne.Size) {
	s := rect.Size()
	return fyne.Position(rect.UL), fyne.Size{Width: s.X, Height: s.Y}
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
func (w *whiteboardRenderer) Refresh() {
	w.shapeLines = w.shapeLines[:0]
	for _, symbol := range w.widget.Content {
		w.shapeLines = append(w.shapeLines, symbolCanvasObjects(symbol)...)
	}
	w.scopesRects = w.scopesRects[:0]
	for _, scope := range w.widget.Scopes {
		r := canvas.NewRectangle(color.RGBA{100, 100, 0, 100})
		pos, size := rectToFyne(scope)
		r.Move(pos)
		r.Resize(size)
		w.scopesRects = append(w.scopesRects, r)
	}
	pos, size := rectToFyne(w.widget.HighlightedScope)
	w.higlightedScope.Move(pos)
	w.higlightedScope.Resize(size)

	canvas.Refresh(w.widget)
}

// Whiteboard is a wigdet displaying a white rectangle, recording
// and displaying one line of user input
type Whiteboard struct {
	// OnEndShape is an optionnal callback
	// triggered after each calls to Rec.EndShape,
	// and before updating the UI.
	OnEndShape func()

	OnCursorMove func(sym.Pos)

	Recorder sym.Recorder

	Content []sym.Symbol

	Scopes []sym.Rect

	HighlightedScope sym.Rect

	widget.BaseWidget
}

func NewWhiteboard() *Whiteboard {
	out := &Whiteboard{}
	out.ExtendBaseWidget(out)
	return out
}

func (w *Whiteboard) CreateRenderer() fyne.WidgetRenderer { return newWhiteboardRendered(w) }

func (w *Whiteboard) MouseDown(*desktop.MouseEvent) {
	w.Recorder.StartShape()
}

func (w *Whiteboard) MouseUp(*desktop.MouseEvent) {
	w.Recorder.EndShape()
	if w.OnEndShape != nil {
		w.OnEndShape()
	}
	w.Refresh()
}

// MouseIn is a hook that is called if the mouse pointer enters the element.
func (w *Whiteboard) MouseIn(*desktop.MouseEvent) {}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (w *Whiteboard) MouseMoved(event *desktop.MouseEvent) {
	w.Recorder.AddToShape(sym.Pos(event.PointEvent.Position))
	if w.OnCursorMove != nil {
		w.OnCursorMove(sym.Pos(event.PointEvent.Position))
		w.Refresh()
	}
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (w *Whiteboard) MouseOut() {
	w.HighlightedScope = sym.Rect{}
	w.Refresh()
}

func (w *Whiteboard) RootScope() sym.Rect {
	s := w.Size()
	return sym.Rect{UL: sym.Pos{X: 0, Y: 0}, LR: sym.Pos{X: s.Width, Y: s.Height}}
}

func shapeCanvasObjects(sh sym.Shape) []fyne.CanvasObject {
	if len(sh) == 0 {
		return nil
	}
	out := make([]fyne.CanvasObject, 0, len(sh)-1)
	start := sh[0]
	for _, point := range sh[1:] {
		li := canvas.NewLine(colornames.Red)

		li.Position1 = fyne.Position(start)
		li.Position2 = fyne.Position(point)
		start = point
		out = append(out, li)
	}
	return out
}

func symbolCanvasObjects(sy sym.Symbol) []fyne.CanvasObject {
	var out []fyne.CanvasObject
	for _, shape := range sy {
		out = append(out, shapeCanvasObjects(shape)...)
	}
	return out
}
