package views

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/benoitkugler/pen2latex/fyneUI/whiteboard"
	"github.com/benoitkugler/pen2latex/layout"
	"github.com/benoitkugler/pen2latex/symbols"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

func showEditor(db *symbols.SymbolStore) *fyne.Container {
	// ed := newEditor(db)

	// data := canvas.NewImageFromImage(image.NewGray(image.Rect(0, -180, 100, 180)))
	// data.FillMode = canvas.ImageFillOriginal

	// shapeImg := canvas.NewImageFromImage(image.NewNRGBA(image.Rect(0, 0, 200, 100)))
	// shapeImg.FillMode = canvas.ImageFillOriginal
	// shapeImg := canvas.NewRasterFromImage(image.NewAlpha(image.Rect(0, 0, 0, 0)))
	rec := whiteboard.NewRecorder()
	rec.OnEndShape = func() {
		// img := rec.Recorder.Current().Shape().AngularGraph()
		// data.Image = img
		// data.Refresh()

		// si := rec.Recorder.Current().Shape().AngleClustersGraph()

		shape := rec.Recorder.Current().Shape()
		fmt.Println("Shape", shape)
		segments := shape.Segment()
		fmt.Println("K", len(segments), segments)
		imag := renderAtoms(segments, shape.BoundingBox())
		savePng(imag)
		// *shapeImg = *canvas.NewRasterFromImage(imag)
		// shapeImg.Resize(fyne.Size{Width: float32(imag.Bounds().Dx()), Height: float32(imag.Bounds().Dy())})
		// shapeImg.Refresh()
	}
	return container.NewVBox(rec)
}

type editor struct {
	wb          *whiteboard.Whiteboard
	recognized  *widget.Label
	resetButton *widget.Button

	db *symbols.SymbolStore

	line layout.Line
}

func newEditor(db *symbols.SymbolStore) *editor {
	ed := &editor{
		wb:          whiteboard.NewWhiteboard(),
		recognized:  widget.NewLabel("Dessiner un caractère..."),
		resetButton: widget.NewButton("Effacer", nil),
		db:          db,
	}
	ed.wb.OnEndShape = ed.tryMatchShape
	ed.wb.OnCursorMove = ed.showScope
	ed.resetButton.OnTapped = ed.clear
	return ed
}

func (ed *editor) tryMatchShape() {
	rec := ed.wb.Recorder.Current()
	if len(rec.Compound()) == 0 {
		return
	}
	ed.line.Insert(rec, ed.db)
	ed.wb.Content = ed.line.Symbols()
	ed.wb.Scopes = ed.line.Scopes()
	ed.recognized.SetText(ed.line.LaTeX())
}

func (ed *editor) showScope(pos symbols.Pos) {
	glyph := symbols.Rect{
		UL: symbols.Pos{X: pos.X - 1, Y: pos.Y - 1},
		LR: symbols.Pos{X: pos.X + 1, Y: pos.Y + 1},
	}
	_, scope, _ := ed.line.FindNode(glyph)
	if scope.IsEmpty() { // root
		scope = ed.wb.RootScope()
	}
	ed.wb.HighlightedScope = scope
}

func (ed *editor) clear() {
	ed.wb.Recorder.Reset()
	ed.line = layout.Line{}
	ed.wb.Content = nil
	ed.wb.Scopes = nil
	ed.wb.Refresh()
	ed.recognized.SetText("Dessiner un caractère...")
}

func posToFixed(pos symbols.Pos) fixed.Point26_6 {
	return fixed.Point26_6{X: fixed.Int26_6(pos.X * 64), Y: fixed.Int26_6(pos.Y * 64)}
}

func renderAtom(atom symbols.ShapeAtom, rec *rasterx.Stroker) {
	switch atom := atom.(type) {
	case symbols.BezierC:
		rec.Start(posToFixed(atom.P0))
		rec.CubeBezier(posToFixed(atom.P1), posToFixed(atom.P2), posToFixed(atom.P3))
		rec.Stop(false)
		// draw control points
		rasterx.AddCircle(float64(atom.P1.X), float64(atom.P1.Y), 3, &rec.Filler)
		rasterx.AddCircle(float64(atom.P2.X), float64(atom.P2.Y), 3, &rec.Filler)
	case symbols.Circle:
		rasterx.AddCircle(float64(atom.Center.X), float64(atom.Center.Y), float64(atom.Radius), rec)
	case symbols.Segment:
		rec.Start(posToFixed(atom.Start))
		rec.Line(posToFixed(atom.End))
		rec.Stop(false)
	}
}

func renderAtoms(atoms []symbols.ShapeAtom, bbox symbols.Rect) image.Image {
	r := image.Rect(int(bbox.LR.X), int(bbox.LR.Y), int(bbox.UL.X), int(bbox.UL.Y))
	bounds := image.Rect(0, 0, r.Max.X, r.Max.Y)
	fmt.Println(bounds)
	img := image.NewNRGBA(bounds)
	sc := rasterx.NewScannerGV(bounds.Dx(), bounds.Dy(), img, bounds)
	rec := rasterx.NewStroker(bounds.Dx(), bounds.Dy(), sc)
	for _, atom := range atoms {
		renderAtom(atom, rec)
	}
	rec.Draw()
	return img
}

func savePng(img image.Image) error {
	out, err := os.Create("tmp.png")
	if err != nil {
		return err
	}

	err = png.Encode(out, img)
	if err != nil {
		return err
	}
	return nil
}
