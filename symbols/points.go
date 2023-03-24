package symbols

import (
	"fmt"
	"math"
	"strings"
)

type Fl = float32

var Inf = Fl(math.Inf(+1))

func Min(x, y Fl) Fl {
	if x < y {
		return x
	}
	return y
}

func Max(x, y Fl) Fl {
	if x > y {
		return x
	}
	return y
}

func abs(x Fl) Fl {
	if x < 0 {
		return -x
	}
	return x
}

func Sqrt(x Fl) Fl { return Fl(math.Sqrt(float64(x))) }

// Pos is a 2D point
type Pos struct {
	X, Y Fl
}

func (p Pos) String() string { return fmt.Sprintf("{X: %.01f, Y:%.01f}", p.X, p.Y) }

func (p *Pos) Scale(s Fl)       { p.X *= s; p.Y *= s }
func (p *Pos) normalize()       { p.Scale(1 / p.Norm()) }
func (p Pos) ScaleTo(s Fl) Pos  { p.Scale(s); return p }
func (p Pos) Add(other Pos) Pos { p.X += other.X; p.Y += other.Y; return p }
func (p Pos) Sub(other Pos) Pos { p.X -= other.X; p.Y -= other.Y; return p }
func (p Pos) Norm() Fl          { return Sqrt(p.NormSquared()) }
func (p Pos) NormSquared() Fl   { return p.X*p.X + p.Y*p.Y }

// return the quadratic norm
func distP(p1, p2 Pos) Fl { return p1.Sub(p2).Norm() }

// returns the dot product of a and b
func dotProduct(a, b Pos) Fl { return (a.X * b.X) + (a.Y * b.Y) }

// return angle in degree
func angle(u, v Pos) Fl {
	dot := float64(u.X*v.X + u.Y*v.Y) // u.v
	det := float64(u.X*v.Y - u.Y*v.X) // u ^ v
	angle := math.Atan2(det, dot)     // in radian
	return Fl(angle * 180 / math.Pi)  // in degre
}

// EmptyRect represents an empty rectangle,
// honoring the following equalities :
//   - r.Union(EmptyRect) == r
//   - EmptyRect.Union(r) == r
//   - EmptyRect.enlarge(p) == p
func EmptyRect() Rect {
	return Rect{
		UL: Pos{X: Inf, Y: Inf},
		LR: Pos{X: -Inf, Y: -Inf},
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
	r.UL.X = Min(r.UL.X, point.X)
	r.UL.Y = Min(r.UL.Y, point.Y)
	r.LR.X = Max(r.LR.X, point.X)
	r.LR.Y = Max(r.LR.Y, point.Y)
}

func (r *Rect) Union(other Rect) {
	r.UL.X = Min(r.UL.X, other.UL.X)
	r.UL.Y = Min(r.UL.Y, other.UL.Y)
	r.LR.X = Max(r.LR.X, other.LR.X)
	r.LR.Y = Max(r.LR.Y, other.LR.Y)
}

func (r Rect) Intersection(other Rect) Rect {
	upperY := Max(r.UL.Y, other.UL.Y)
	lowerY := Min(r.LR.Y, other.LR.Y)
	leftX := Max(r.UL.X, other.UL.X)
	rightX := Min(r.LR.X, other.LR.X)
	if upperY > lowerY || leftX > rightX {
		return EmptyRect()
	}
	return Rect{
		UL: Pos{X: leftX, Y: upperY},
		LR: Pos{X: rightX, Y: lowerY},
	}
}

func (r Rect) Area() Fl { return r.Width() * r.Height() }

func (r Rect) Size() Pos { return Pos{r.Width(), r.Height()} }

func (r Rect) Width() Fl  { return Max(r.LR.X-r.UL.X, 0) }
func (r Rect) Height() Fl { return Max(r.LR.Y-r.UL.Y, 0) }

// Shape stores the points of a shape drawn without lifting the pen
type Shape []Pos

func (sh Shape) String() string {
	chunks := make([]string, len(sh))
	for i, p := range sh {
		chunks[i] = p.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(chunks, ","))
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

func (sy Symbol) IsCompound() bool { return len(sy) > 1 }

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
		out = append(out, seg.p0.Add(AB.ScaleTo(Fl(i)/nbPoints)))
	}
	return out
}
