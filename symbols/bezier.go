package symbols

import (
	"fmt"
	"math"
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
	db := b.derivativeAt(t)
	ddb := b.secondDerivativeAt(t)

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

func (be Bezier) boundingBox() Rect {
	re := EmptyRect()
	for _, p := range be.toPoints() {
		re.enlarge(p)
	}
	return re
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

// after clippping return the approximated line portions that intersects
// taken from https://stackoverflow.com/a/4041286
func (U Bezier) HasIntersection(other Bezier) bool {
	const threshold = 0.1
	var aux func(b1, b2 Bezier) bool
	aux = func(b1, b2 Bezier) bool {
		bbox1, bbox2 := b1.controlBox(), b2.controlBox()

		if bbox1.Intersection(bbox2).IsEmpty() {
			return false
		}

		a1, a2 := b1.arcLength(), b2.arcLength()
		if a1+a2 < threshold { // found one intersection
			return true
		}

		// recurse
		b11, b12 := b1.splitAt(0.5)
		b21, b22 := b2.splitAt(0.5)
		return aux(b11, b21) ||
			aux(b11, b22) ||
			aux(b12, b21) ||
			aux(b12, b22)
	}

	return aux(U, other)
}

// returns true if the curve describes one point
func (U Bezier) isAlmostPoint() (Pos, bool) {
	barycentre := U.P0.Add(U.P3).ScaleTo(0.5)

	cb := U.controlBox()

	return barycentre, cb.Width() <= 1 && cb.Height() <= 1
}

// IsPoint return true if the curve describes a single point
func (U Bezier) IsPoint() (Pos, bool) {
	return U.P0, U.P0 == U.P1 && U.P1 == U.P2 && U.P2 == U.P3
}

// IsRoughlyLinear returns true for curves that could
// be approximated by a segment, for collisions purposes
func (U Bezier) IsRoughlyLinear() bool {
	return U.diffWithLine() < 0.1
}

// assume control points are inside the edges
// return a normalized value, tpycally under 0.1 for lines
// returns Inf if it can't be a line
func (U Bezier) diffWithLine() Fl {
	U.normalize()

	if U.P1.X < U.P1.X || U.P2.X > U.P3.X {
		return Inf
	}

	dx := U.P3.X
	// normalize width to 1 as well
	U = U.Scale(Trans{Scale: 1 / dx})

	points := U.toPoints()
	var area Fl
	for _, p := range points {
		area += abs(p.Y)
	}
	area /= Fl(len(points))

	return area
}

func areTangentsAligned(c1, c2 Bezier) bool {
	d1 := c1.P3.Sub(c1.P2)
	d2 := c2.P1.Sub(c2.P0)
	return abs(angle(d1, d2)) < 5
}

// translate and rotate so that P0 = 0, P3.Y = 0, P3.X > 0
func (b *Bezier) normalize() (center Pos, c, s Fl) {
	// shift to (0,0)
	center = b.P0
	b.P0 = b.P0.Sub(center)
	b.P1 = b.P1.Sub(center)
	b.P2 = b.P2.Sub(center)
	b.P3 = b.P3.Sub(center)

	// rotate
	theta := math.Atan2(float64(b.P3.Y), float64(b.P3.X))
	c, s = Fl(math.Cos(-theta)), Fl(math.Sin(-theta))
	b.P1.rotate(c, s)
	b.P2.rotate(c, s)
	b.P3.rotate(c, s)

	return
}

// return roots in [0,1]
func quadraticRoots(a, b, c Fl) []Fl {
	if a == 0 {
		t := -c / b
		return []Fl{t}
	}

	delta := b*b - 4*a*c
	if delta < 0 {
		return nil
	}

	sd := Sqrt(delta)
	t1 := (-b + sd) / (2 * a)
	t2 := (-b - sd) / (2 * a)
	var out []Fl
	if 0 <= t1 && t1 <= 1 {
		out = append(out, t1)
	}
	if 0 <= t2 && t2 <= 1 && t2 != t1 {
		out = append(out, t2)
	}

	return out
}

func (b Bezier) criticalPointsY() []Fl {
	p0, p1, p2, p3 := b.P0, b.P1, b.P2, b.P3
	q0, q1, q2 := p1.Sub(p0), p2.Sub(p1), p3.Sub(p2)
	A := q0.Sub(q1.ScaleTo(2)).Add(q2)
	B := q1.Sub(q0).ScaleTo(2)
	C := q0

	return quadraticRoots(A.Y, B.Y, C.Y)
}

// hasRoughEndAngle returns true if be ends with a rough angle,
// which should be split
func (be Bezier) hasRoughEndAngle() (Fl, bool) {
	be.normalize()

	// controls must be opposite, with P2 "after" P1 and P3
	if !(be.P2.X > be.P3.X && be.P1.Y*be.P2.Y < 0) {
		return 0, false
	}

	ts := be.criticalPointsY()
	if len(ts) == 0 {
		return 0, false
	}
	t := ts[0]
	if len(ts) == 2 {
		t = Max(t, ts[1])
	}

	ref := be.P0.Sub(be.P3).NormSquared()
	curvature := be.curvatureAt(t) // evolves as 1/ \lambda^2
	curvatureNormalized := abs(curvature * ref)
	const threshold = 1000

	if curvatureNormalized > threshold {
		return t, true
	}

	return 0, false
}
