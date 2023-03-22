package whiteboard

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/layout"
	sym "github.com/benoitkugler/pen2latex/symbols"
	"golang.org/x/image/colornames"
)

var (
	_ desktop.Mouseable = (*Recorder)(nil)
	_ desktop.Hoverable = (*Recorder)(nil)

	_ fyne.WidgetRenderer = (*recorderRenderer)(nil)
)

type recorderRenderer struct {
	widget *Recorder

	rec        *canvas.Rectangle
	baseline   *canvas.Line
	shapeLines []fyne.CanvasObject
}

func newRecorderREnderer(w *Recorder) *recorderRenderer {
	return &recorderRenderer{
		widget:   w,
		rec:      canvas.NewRectangle(colornames.Lightgray),
		baseline: canvas.NewLine(color.Black),
	}
}

func (w *recorderRenderer) Destroy() {}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
func (w *recorderRenderer) Layout(size fyne.Size) {
	w.rec.Resize(size)
	w.baseline.Position1 = fyne.NewPos(0, size.Height*sym.EMBaselineRatio)
	w.baseline.Position2 = fyne.NewPos(size.Width, size.Height*sym.EMBaselineRatio)
}

func (w *recorderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(sym.EMWidth, 5*sym.EMHeight)
}

// Objects returns all objects that should be drawn.
func (w *recorderRenderer) Objects() []fyne.CanvasObject {
	out := make([]fyne.CanvasObject, 0, 2+len(w.shapeLines))
	out = append(out, w.rec, w.baseline)
	out = append(out, w.shapeLines...)
	return out
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
func (w *recorderRenderer) Refresh() {
	w.shapeLines = w.shapeLines[:0]
	for _, symbol := range w.widget.Content {
		w.shapeLines = append(w.shapeLines, symbolCanvasObjects(symbol)...)
	}
	canvas.Refresh(w.widget)
}

type record struct {
	ti  time.Time
	pos sym.Pos
}

// Recorder is a wigdet displaying a white rectangle, recording
// and displaying one line of user input
type Recorder struct {
	OnEndShape func()

	Points []record

	Recorder layout.Recorder

	Content []sym.Symbol

	Scopes []sym.Rect

	HighlightedScope sym.Rect

	widget.BaseWidget
}

func NewRecorder() *Recorder {
	out := &Recorder{}
	out.ExtendBaseWidget(out)
	return out
}

func (w *Recorder) CreateRenderer() fyne.WidgetRenderer { return newRecorderREnderer(w) }

func (w *Recorder) MouseDown(*desktop.MouseEvent) {
	w.Recorder.StartShape()
	w.Points = nil
}

func (w *Recorder) MouseUp(*desktop.MouseEvent) {
	w.Recorder.EndShape()
	if w.OnEndShape != nil {
		w.OnEndShape()
	}
	w.Refresh()
}

// MouseIn is a hook that is called if the mouse pointer enters the element.
func (w *Recorder) MouseIn(*desktop.MouseEvent) {}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (w *Recorder) MouseMoved(event *desktop.MouseEvent) {
	pos := sym.Pos(event.PointEvent.Position)
	w.Recorder.AddToShape(pos)
	w.Points = append(w.Points, record{ti: time.Now(), pos: pos})
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (w *Recorder) MouseOut() {}
