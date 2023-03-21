package symbols

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type fl = float32

func min(x, y fl) fl {
	if x < y {
		return x
	}
	return y
}

func max(x, y fl) fl {
	if x > y {
		return x
	}
	return y
}

func abs(x fl) fl {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x fl) fl { return fl(math.Sqrt(float64(x))) }

// Pos is a 2D point
type Pos struct {
	X, Y fl
}

func (p Pos) String() string { return fmt.Sprintf("{X: %.01f, Y:%.01f}", p.X, p.Y) }

func (p *Pos) Scale(s fl)       { p.X *= s; p.Y *= s }
func (p *Pos) normalize()       { p.Scale(1 / p.Norm()) }
func (p Pos) ScaleTo(s fl) Pos  { p.Scale(s); return p }
func (p Pos) Add(other Pos) Pos { p.X += other.X; p.Y += other.Y; return p }
func (p Pos) Sub(other Pos) Pos { p.X -= other.X; p.Y -= other.Y; return p }
func (p Pos) Norm() fl          { return sqrt(p.NormSquared()) }
func (p Pos) NormSquared() fl   { return p.X*p.X + p.Y*p.Y }

// return the quadratic norm
func distP(p1, p2 Pos) fl { return p1.Sub(p2).Norm() }

// returns the dot product of a and b
func dotProduct(a, b Pos) fl { return (a.X * b.X) + (a.Y * b.Y) }

var inf = fl(math.Inf(+1))

// return angle in degree
func angle(u, v Pos) fl {
	dot := float64(u.X*v.X + u.Y*v.Y) // u.v
	det := float64(u.X*v.Y - u.Y*v.X) // u ^ v
	angle := math.Atan2(det, dot)     // in radian
	return fl(angle * 180 / math.Pi)  // in degre
}

// EmptyRect represents an empty rectangle,
// honoring the following equalities :
//   - r.Union(EmptyRect) == r
//   - EmptyRect.Union(r) == r
//   - EmptyRect.enlarge(p) == p
func EmptyRect() Rect {
	return Rect{
		UL: Pos{X: inf, Y: inf},
		LR: Pos{X: -inf, Y: -inf},
	}
}

// Rect is a 2D rectangle, defined by
// its upper left and lower right corner.
type Rect struct {
	UL, LR Pos
}

func (r Rect) String() string {
	return fmt.Sprintf("{UL: %s, LR: %s}", r.UL, r.LR)
}

func (r Rect) IsEmpty() bool { return r == EmptyRect() }

func (r Rect) contains(p Pos) bool {
	if r.IsEmpty() {
		return false
	}
	return r.UL.X <= p.X && p.X <= r.LR.X && r.UL.Y <= p.Y && p.Y <= r.LR.Y
}

func (r Rect) tranlate(p Pos) Rect {
	r.UL = r.UL.Add(p)
	r.LR = r.LR.Add(p)
	return r
}

func (r *Rect) enlarge(point Pos) {
	r.UL.X = min(r.UL.X, point.X)
	r.UL.Y = min(r.UL.Y, point.Y)
	r.LR.X = max(r.LR.X, point.X)
	r.LR.Y = max(r.LR.Y, point.Y)
}

func (r *Rect) Union(other Rect) {
	r.UL.X = min(r.UL.X, other.UL.X)
	r.UL.Y = min(r.UL.Y, other.UL.Y)
	r.LR.X = max(r.LR.X, other.LR.X)
	r.LR.Y = max(r.LR.Y, other.LR.Y)
}

func (r Rect) Intersection(other Rect) Rect {
	upperY := max(r.UL.Y, other.UL.Y)
	lowerY := min(r.LR.Y, other.LR.Y)
	leftX := max(r.UL.X, other.UL.X)
	rightX := min(r.LR.X, other.LR.X)
	if upperY > lowerY || leftX > rightX {
		return EmptyRect()
	}
	return Rect{
		UL: Pos{X: leftX, Y: upperY},
		LR: Pos{X: rightX, Y: lowerY},
	}
}

func (r Rect) Area() fl { return r.Width() * r.Height() }

func (r Rect) Size() Pos { return Pos{r.Width(), r.Height()} }

func (r Rect) Width() fl  { return max(r.LR.X-r.UL.X, 0) }
func (r Rect) Height() fl { return max(r.LR.Y-r.UL.Y, 0) }

// Shape stores the points of a shape drawn without lifting the pen
type Shape []Pos

func (sh Shape) String() string {
	var st strings.Builder
	st.WriteString("Shape{\n")
	for _, p := range sh {
		st.WriteString(p.String())
		st.WriteString(",")
	}
	st.WriteString("}\n")
	return st.String()
}

// BoundingBox returns the rectangle enclosing the shape.
// It returns EmptyRect if the shape is empty
func (sh Shape) BoundingBox() Rect {
	out := EmptyRect()
	for _, point := range sh {
		out.enlarge(point)
	}
	return out
}

// Symbol is an union of connex shapes,
// used to represent one grapheme (like Sigma, an accentued character, etc...)
type Symbol []Shape

// Union merge all the connex components
func (sy Symbol) Union() Shape {
	var out Shape
	for _, r := range sy {
		out = append(out, r...)
	}
	return out
}

func (seg segment) toPoints() Shape {
	const nbPoints = 20
	var out Shape
	AB := seg.p1.Sub(seg.p0)
	for i := 0; i < nbPoints; i++ {
		out = append(out, seg.p0.Add(AB.ScaleTo(fl(i)/nbPoints)))
	}
	return out
}

type record struct {
	shape  Shape
	timing int64 // in milliseconds
}

type Record struct {
	allShapes []Shape
}

// Compound returns the whole symbol currently drawn
func (rec Record) Compound() Symbol { return rec.allShapes }

// Shape returns the last connex compoment drawn
func (rec Record) Shape() Shape {
	if len(rec.allShapes) == 0 {
		return nil
	}
	return rec.allShapes[len(rec.allShapes)-1]
}

// LastCompound returns the symbol drawn before the last connex shape
func (rec Record) LastCompound() Symbol {
	if len(rec.allShapes) == 0 {
		return nil
	}
	return rec.allShapes[0 : len(rec.allShapes)-1]
}

// Recorder is used to record the shapes
// drawn by the user
type Recorder struct {
	// trace is the accumulated trace
	trace []record

	currentShape Shape
	inShape      bool
}

// Reset clears the current state of the `Recorder`
func (rec *Recorder) Reset() {
	rec.inShape = false
	rec.trace = rec.trace[:0]
	rec.currentShape = nil
}

// Symbol returns the current symbol being recorded.
// More precisely, it returns the last [Shape] drawn,
// and the whole compound symbol.
func (rec Recorder) Current() Record {
	var out Symbol
	for _, r := range rec.trace {
		out = append(out, r.shape)
	}
	return Record{out}
}

// StartShape starts the recording of a new connex shape.
func (rec *Recorder) StartShape() {
	rec.inShape = true
	rec.currentShape = Shape{}
}

// EndShape ends the recording of the current connex shape.
// Empty shapes are discarded.
func (rec *Recorder) EndShape() {
	rec.inShape = false
	if current := rec.currentShape; len(current) != 0 {
		rec.trace = append(rec.trace, record{rec.currentShape, time.Now().UnixMilli()})
	}
}

// AddToShape adds a point to the shape.
func (rec *Recorder) AddToShape(pos Pos) {
	if rec.inShape {
		rec.currentShape = append(rec.currentShape, pos)
	}
}
