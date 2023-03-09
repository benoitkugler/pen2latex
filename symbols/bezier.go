package symbols

import (
	"math"
)

type bezierL struct {
	P0, P1 Pos
}

func (b bezierL) eval(t fl) (out Pos) {
	t1 := 1 - t
	A := t1
	B := t
	out.X = A*b.P0.X + B*b.P1.X
	out.Y = A*b.P0.Y + B*b.P1.Y
	return
}

type bezierQ struct {
	P0, P1, P2 Pos
}

func (b bezierQ) eval(t fl) (out Pos) {
	t1 := 1 - t
	A := t1 * t1
	B := 2 * t1 * t
	C := t * t
	out.X = A*b.P0.X + B*b.P1.X + C*b.P2.X
	out.Y = A*b.P0.Y + B*b.P1.Y + C*b.P2.Y
	return
}

// Bezier is a cubic Bezier curve
type Bezier struct {
	P0, P1, P2, P3 Pos
}

func (b Bezier) eval(t fl) (out Pos) {
	t1 := 1 - t
	A := t1 * t1 * t1
	B := 3 * t1 * t1 * t
	C := 3 * t1 * t * t
	D := t * t * t
	out.X = A*b.P0.X + B*b.P1.X + C*b.P2.X + D*b.P3.X
	out.Y = A*b.P0.Y + B*b.P1.Y + C*b.P2.Y + D*b.P3.Y
	return
}

func (knots Shape) ToBezierCurves() []Bezier {
	if len(knots) <= 1 {
		return nil
	}

	n := len(knots) - 1
	if n == 1 { // Special case: Bezier curve should be a straight line.
		var P1, P2 Pos
		// 3P1 = 2P0 + P3
		P1.X = (2*knots[0].X + knots[1].X) / 3
		P1.Y = (2*knots[0].Y + knots[1].Y) / 3

		// P2 = 2P1 – P0
		P2.X = 2*P1.X - knots[0].X
		P2.Y = 2*P1.Y - knots[0].Y
		return []Bezier{{P0: knots[0], P1: P1, P2: P2, P3: knots[1]}}
	}

	// Calculate first Bezier control points
	// Right hand side vector
	rhs := make([]fl, n)

	// Set right hand side X values
	for i := 1; i < n-1; i++ {
		rhs[i] = 4*knots[i].X + 2*knots[i+1].X
	}
	rhs[0] = knots[0].X + 2*knots[1].X
	rhs[n-1] = (8*knots[n-1].X + knots[n].X) / 2.0
	// Get first control points X-values
	x := getFirstControlPoints(rhs)

	// Set right hand side Y values
	for i := 1; i < n-1; i++ {
		rhs[i] = 4*knots[i].Y + 2*knots[i+1].Y
	}
	rhs[0] = knots[0].Y + 2*knots[1].Y
	rhs[n-1] = (8*knots[n-1].Y + knots[n].Y) / 2.0
	// Get first control points Y-values
	y := getFirstControlPoints(rhs)

	// Fill output arrays.
	out := make([]Bezier, n)
	for i := range out {
		// First control point
		P1 := Pos{x[i], y[i]}
		// Second control point
		var P2 Pos
		if i < n-1 {
			P2 = Pos{2*knots[i+1].X - x[i+1], 2*knots[i+1].Y - y[i+1]}
		} else {
			P2 = Pos{(knots[n].X + x[n-1]) / 2, (knots[n].Y + y[n-1]) / 2}
		}
		out[i] = Bezier{P0: knots[i], P1: P1, P2: P2, P3: knots[i+1]}
	}
	return out
}

// Solves a tridiagonal system for one of coordinates (x or y)
// of first Bezier control points.
func getFirstControlPoints(rhs []fl) []float32 {
	N := len(rhs)
	x := make([]fl, N)   // Solution vector.
	tmp := make([]fl, N) // Temp workspace.

	var b fl = 2.0
	x[0] = rhs[0] / b
	for i := 1; i < N; i++ { // Decomposition and forward substitution.
		tmp[i] = 1 / b
		if i < N-1 {
			b = 4 - tmp[i]
		} else {
			b = 3.5 - tmp[i]
		}
		x[i] = (rhs[i] - x[i-1]) / b
	}
	for i := 1; i < N; i++ { // Backsubstitution.
		x[N-i-1] -= tmp[N-i] * x[N-i]
	}

	return x
}

// translate and rotate so that P0 = 0, P3.Y = 0, P3.X > 0
func (b *Bezier) normalize() (center Pos, c, s fl) {
	// shift to (0,0)
	center = b.P0
	b.P0 = b.P0.Sub(center)
	b.P1 = b.P1.Sub(center)
	b.P2 = b.P2.Sub(center)
	b.P3 = b.P3.Sub(center)

	// rotate
	theta := math.Atan2(float64(b.P3.Y), float64(b.P3.X))
	c, s = fl(math.Cos(-theta)), fl(math.Sin(-theta))
	b.P1.rotate(c, s)
	b.P2.rotate(c, s)
	b.P3.rotate(c, s)

	return
}

// assume b is normalized
func (be Bezier) extremumY() Pos {
	a := be.P1.Y
	b := be.P2.Y - be.P1.Y
	c := be.P2.Y
	A := a - 2*b - c

	if A == 0 {
		t := -a / (2 * (b - a))
		return be.eval(t)
	}

	delta := (b-a)*(b-a) - a*A

	if delta < 0 {
		return Pos{}
	}

	sd := sqrt(delta)
	t1 := ((a - b) + sd) / A
	t2 := ((a - b) - sd) / A
	if 0 <= t1 && t1 <= 1 {
		return be.eval(t1)
	} else if 0 <= t2 && t2 <= 1 {
		return be.eval(t2)
	}
	return Pos{}
}

// assume [be] is fitted from origin
func (be Bezier) extremalPoint(origin Shape) int {
	center, c, s := be.normalize()
	max := be.extremumY()
	var (
		bestIndex int
		bestDist  = inf
	)
	for i, p := range origin {
		p = p.Sub(center)
		p.rotate(c, s)
		if d := p.Sub(max).NormSquared(); d < bestDist {
			bestIndex = i
			bestDist = d
		}
	}
	return bestIndex
}
