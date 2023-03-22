package symbols

import (
	"fmt"
)

// Bezier is a cubic Bezier curve
type Bezier struct {
	P0, P1, P2, P3 Pos
}

func (b Bezier) String() string {
	return fmt.Sprintf("{Pos%s, Pos%s, Pos%s, Pos%s}", b.P0, b.P1, b.P2, b.P3)
}

func (b Bezier) pointAt(t Fl) (out Pos) {
	t1 := 1 - t
	A := t1 * t1 * t1
	B := 3 * t1 * t1 * t
	C := 3 * t1 * t * t
	D := t * t * t
	out.X = A*b.P0.X + B*b.P1.X + C*b.P2.X + D*b.P3.X
	out.Y = A*b.P0.Y + B*b.P1.Y + C*b.P2.Y + D*b.P3.Y
	return
}

// returns B'(t)
func (b Bezier) derivativeAt(t Fl) Pos {
	p0, p1, p2, p3 := b.P0, b.P1, b.P2, b.P3
	q0, q1, q2 := p1.Sub(p0), p2.Sub(p1), p3.Sub(p2)
	return q0.ScaleTo(3 * (1 - t) * (1 - t)).
		Add(q1.ScaleTo(6 * t * (1 - t))).
		Add(q2.ScaleTo(3 * t * t))
}

// returns B”(t)
func (b Bezier) secondDerivativeAt(t Fl) Pos {
	p0, p1, p2, p3 := b.P0, b.P1, b.P2, b.P3
	q0, q1, q2 := p1.Sub(p0), p2.Sub(p1), p3.Sub(p2)
	r0, r1 := q1.Sub(q0), q2.Sub(q1)
	return r0.ScaleTo(6 * (1 - t)).
		Add(r1.ScaleTo(6 * t))
}

func (b Bezier) curvatureAt(t Fl) Fl {
	// k = (x'' y'  - y''x') / (x'^2 + y'^2)^(3/2)
	p0, p1, p2, p3 := b.P0, b.P1, b.P2, b.P3
	db := p1.Sub(p0).ScaleTo(3 * (1 - t) * (1 - t)).
		Add(p2.Sub(p1).ScaleTo(6 * t * (1 - t))).
		Add(p3.Sub(p2).ScaleTo(3 * t * t))
	ddb := p2.Sub(p1.ScaleTo(2)).Add(p0).ScaleTo(6 * (1 - t)).
		Add(p3.Sub(p2.ScaleTo(2)).Add(p1).ScaleTo(6 * t))

	denom := db.Norm()
	denom = denom * denom * denom

	return (ddb.X*db.Y - ddb.Y*db.X) / denom
}

// De Casteljau's algorithm
func (b Bezier) splitAt(t Fl) (b1, b2 Bezier) {
	if t == 0 {
		return Bezier{}, b
	}
	if t == 1 {
		return b, Bezier{}
	}

	// see also https://stackoverflow.com/questions/18655135/divide-bezier-curve-into-two-equal-halves
	A, B, C, D := b.P0, b.P1, b.P2, b.P3

	E := A.ScaleTo(1 - t).Add(B.ScaleTo(t))
	F := B.ScaleTo(1 - t).Add(C.ScaleTo(t))
	G := C.ScaleTo(1 - t).Add(D.ScaleTo(t))

	H := E.ScaleTo(1 - t).Add(F.ScaleTo(t))
	J := F.ScaleTo(1 - t).Add(G.ScaleTo(t))

	K := H.ScaleTo(1 - t).Add(J.ScaleTo(t))

	return Bezier{A, E, H, K}, Bezier{K, J, G, D}
}

func (be Bezier) controlBox() Rect {
	re := Rect{be.P0, be.P0}
	re.enlarge(be.P1)
	re.enlarge(be.P2)
	re.enlarge(be.P3)
	return re
}

func (b Bezier) splitBetween(t0, t1 Fl) Bezier {
	// first split
	_, right := b.splitAt(t0)
	// convert t1 from [t0, 1] to [0,1]
	t1 = (t1 - t0) / (1 - t0)
	center, _ := right.splitAt(t1)
	return center
}

func (b Bezier) toPoints() Shape {
	const nbPoints = 40
	var sh Shape
	for i := 0; i < nbPoints; i++ {
		sh = append(sh, b.pointAt(Fl(i)/nbPoints))
	}
	return sh
}

func (b Bezier) arcLength() Fl {
	const nbPoints = 100
	previousPos := b.P0
	var length Fl
	for i := 1; i < nbPoints; i++ {
		pos := b.pointAt(Fl(i) / nbPoints)
		length += distP(pos, previousPos)
		previousPos = pos
	}
	return length
}
