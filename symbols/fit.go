package symbols

import (
	"math"
)

// fitting algorithm for "elementary" shapes

/*
An Algorithm for Automatically Fitting Digitized Curves
by Philip J. Schneider
from "Graphics Gems", Academic Press, 1990
*/

// panic if len(d) < 3
// return the average quadratic error
func fitCubicBezier(points []Pos) (BezierC, fl) {
	const maxIterations = 8 // tuned experimentally

	// unit tangent vectors at endpoints
	tHat1 := computeLeftTangent(points, 0)
	tHat2 := computeRightTangent(points, len(points)-1)

	bestErr, bestBezier := inf, BezierC{}

	// start parametrization with path arc length
	u := pathLengthIndices(points)

	for i := 0; i < maxIterations; i++ {
		bezCurve := generateBezier(points, u, tHat1, tHat2)
		errValue := computeBezierDistance(points, bezCurve, u)
		if errValue < bestErr {
			bestErr = errValue
			bestBezier = bezCurve
		}

		u = reparameterize(points, u, bezCurve)
	}
	return bestBezier, bestErr
}

// return t_i : coefficients in [0, 1], computed from the path length
func pathLengthIndices(points []Pos) []fl {
	out := make([]fl, len(points))
	var totalLength fl
	for i := range points {
		if i == 0 {
			continue
		}
		totalLength += distancePoints(points[i], points[i-1])
		out[i] = totalLength
	}
	// normalize to [0,1]
	for i := range out {
		out[i] /= totalLength
	}
	return out
}

// u is the current parametrization of the points
func generateBezier(d []Pos, u []fl, tHat1, tHat2 Pos) BezierC {
	first, last := 0, len(d)-1

	// Compute A, rhs for eqn
	A := make([][2]Pos, len(d))
	for i, ui := range u {
		v1 := tHat1
		v2 := tHat2
		v1.Scale(3 * ui * (1 - ui) * (1 - ui))
		v2.Scale(3 * ui * ui * (1 - ui))
		A[i][0] = v1
		A[i][1] = v2
	}

	/* Create the C and X matrices  */
	var (
		C [2][2]fl
		X [2]fl
	)

	for i := range d {
		C[0][0] += dotProduct(A[i][0], A[i][0])
		C[0][1] += dotProduct(A[i][0], A[i][1])
		C[1][0] = C[0][1]
		C[1][1] += dotProduct(A[i][1], A[i][1])

		bez := BezierC{d[first], d[first], d[last], d[last]}.eval(u[i])

		tmp := d[first+i].Sub(bez)

		X[0] += dotProduct(A[i][0], tmp)
		X[1] += dotProduct(A[i][1], tmp)
	}

	/* Compute the determinants of C and X  */
	det_C0_C1 := C[0][0]*C[1][1] - C[1][0]*C[0][1]
	det_C0_X := C[0][0]*X[1] - C[1][0]*X[0]
	det_X_C1 := X[0]*C[1][1] - X[1]*C[0][1]

	/* Finally, derive alpha values */
	var alpha_l, alpha_r fl
	if det_C0_C1 != 0 {
		alpha_l = det_X_C1 / det_C0_C1
	}
	if det_C0_C1 != 0 {
		alpha_r = det_C0_X / det_C0_C1
	}

	/* If alpha negative, use the Wu/Barsky heuristic (see text) */
	/* (if alpha is 0, you get coincident control points that lead to
	 * divide by zero in any subsequent newtonRaphsonRootStep() call. */
	segLength := distancePoints(d[first], d[last])
	epsilon := 1.0e-6 * segLength
	if alpha_l < epsilon || alpha_r < epsilon {
		/* fall back on standard (probably inaccurate) formula, and subdivide further if needed. */
		dist := segLength / 3.0
		return BezierC{
			P0: d[first],
			P1: tHat1.ScaleTo(dist).Add(d[first]),
			P2: tHat2.ScaleTo(dist).Add(d[last]),
			P3: d[last],
		}
	}

	/*  First and last control points of the Bezier curve are */
	/*  positioned exactly at the first and last data points */
	/*  Control points 1 and 2 are positioned an alpha distance out */
	/*  on the tangent vectors, left and right, respectively */
	return BezierC{
		P0: d[first],
		P1: tHat1.ScaleTo(alpha_l).Add(d[first]),
		P2: tHat2.ScaleTo(alpha_r).Add(d[last]),
		P3: d[last],
	}
}

// Given set of points and their parameterization, try to find
// a better parameterization.
func reparameterize(d []Pos, u []fl, bezCurve BezierC) []fl {
	uPrime := make([]fl, len(d)) //  new parameter values
	for i := range d {
		uPrime[i] = newtonRaphsonRootStep(bezCurve, d[i], u[i])
	}
	return uPrime
}

// newtonRaphsonRootStep :
// Use Newton-Raphson iteration to find better root.
func newtonRaphsonRootStep(Q BezierC, P Pos, u fl) fl {
	// Compute Q(u)
	Q_u := Q.eval(u)

	/*  Q' and Q''          */
	Q1 := bezierQ{
		P0: Q.P1.Sub(Q.P0).ScaleTo(3),
		P1: Q.P2.Sub(Q.P1).ScaleTo(3),
		P2: Q.P3.Sub(Q.P2).ScaleTo(3),
	}
	Q2 := bezierL{
		P0: Q.P1.Sub(Q.P0).ScaleTo(2),
		P1: Q.P2.Sub(Q.P1).ScaleTo(2),
	}

	/* Compute Q'(u) and Q''(u) */
	Q1_u := Q1.eval(u)
	Q2_u := Q2.eval(u)

	/* Compute f(u)/f'(u) */
	numerator := (Q_u.X-P.X)*(Q1_u.X) + (Q_u.Y-P.Y)*(Q1_u.Y)
	denominator := (Q1_u.X)*(Q1_u.X) + (Q1_u.Y)*(Q1_u.Y) +
		(Q_u.X-P.X)*(Q2_u.X) + (Q_u.Y-P.Y)*(Q2_u.Y)
	if denominator == 0.0 {
		return u
	}

	/* u = u - f(u)/f'(u) */
	uPrime := u - (numerator / denominator)
	return uPrime
}

// computeLeftTangent, ComputeRightTangent, ComputeCenterTangent :
// Approximate unit tangents at endpoints and "center" of digitized curve
func computeLeftTangent(d []Pos, end int) Pos {
	tHat1 := d[end+1].Sub(d[end])
	tHat1.Scale(tHat1.Norm())
	return tHat1
}

func computeRightTangent(d []Pos, end int) Pos {
	tHat2 := d[end-1].Sub(d[end])
	tHat2.Scale(tHat2.Norm())
	return tHat2
}

// computeDistance finds the average squared distance of digitized points
// to fitted curve, given a parametrization
func computeBezierDistance(d []Pos, bezCurve BezierC, u []fl) fl {
	var average fl // maximum fitError
	for i := range d {
		P := bezCurve.eval(u[i]) /*  Pos on curve      */
		dist := P.Sub(d[i]).NormSquared()
		average += dist
	}
	return average / fl(len(d))
}

// returns the dot product of a and b
func dotProduct(a, b Pos) fl { return (a.X * b.X) + (a.Y * b.Y) }

// Adapted from https://gist.github.com/WetHat/ab410a300cc477f6c2002e0a281cf5b1

type Circle struct {
	Center Pos
	Radius fl
}

func fitCircle(points []Pos) (Circle, fl) {
	var tx, ty, txy fl    // x^2, y^2, xy
	var s0, s1, s2, s3 fl // x^3, x^2y, xy^2, y^3

	// compute barycentre
	var barycentre Pos
	for _, p := range points {
		barycentre.X += p.X
		barycentre.Y += p.Y
	}
	barycentre.Scale(1 / fl(len(points)))

	for _, p := range points {
		x, y := p.X-barycentre.X, p.Y-barycentre.Y
		x2, y2 := x*x, y*y

		tx += x2
		ty += y2
		txy += x * y

		s0 += x * x2
		s1 += x2 * y
		s2 += x * y2
		s3 += y2 * y
	}
	denom := 2 * (tx*ty - txy*txy)
	cx := ((s2+s0)*ty - (s1+s3)*txy) / denom
	cy := ((s1+s3)*tx - (s2+s0)*txy) / denom
	center := Pos{cx, cy}
	r := fl(math.Sqrt(float64(cx*cx + cy*cy + (tx+ty)/fl(len(points)))))

	// translate by the barycentre
	center = center.Add(barycentre)

	// error value
	var averageErr fl
	for _, p := range points {
		tmp := p.Sub(center).Norm() - r // maybe be negative
		averageErr += tmp * tmp
	}
	return Circle{center, r}, averageErr / fl(len(points))
}

// line fit

type Segment struct {
	Start, End Pos
}

func fitSegment(points []Pos) (Segment, fl) {
	// find the line and the find start and end
	var sx, sy, sxy, sx2 fl
	for _, p := range points {
		sx += p.X
		sy += p.Y
		sxy += p.X * p.Y
		sx2 += p.X * p.X
	}
	N := fl(len(points))
	// y = mx + b
	m := (N*sxy - sx*sy) / (N*sx2 - sx*sx)
	b := (sy - m*sx) / N
	// vector director TODO: handle vertical line
	u := Pos{1, m}
	nu := u.NormSquared()
	A := Pos{0, b}

	minT, maxT := inf, -inf
	var (
		start, end Pos
		err        fl
	)
	for _, p := range points {
		t := dotProduct(u, p.Sub(A)) / nu
		H := A.Add(u.ScaleTo(t))
		errValue := p.Sub(H).NormSquared()
		err += errValue
		if t < minT {
			minT = t
			start = H
		}
		if t > maxT {
			maxT = t
			end = H
		}
	}

	return Segment{start, end}, err / N
}

type ShapeAtomKind uint8

const (
	SAKBezier ShapeAtomKind = iota
	SAKSegment
	SAKCircle
)

type ShapeAtom interface {
	Kind() ShapeAtomKind
}

func (BezierC) Kind() ShapeAtomKind { return SAKBezier }
func (Segment) Kind() ShapeAtomKind { return SAKSegment }
func (Circle) Kind() ShapeAtomKind  { return SAKCircle }

func (sh Shape) identify() ShapeAtom {
	if len(sh) == 0 {
		return BezierC{}
	} else if len(sh) <= 2 {
		start, end := sh[0], sh[len(sh)-1]
		return Segment{start, end}
	}

	bezier, errBezier := fitCubicBezier(sh)
	segment, errSegment := fitSegment(sh)
	circle, errCircle := fitCircle(sh)

	// Adjust the raw error with heuristics

	// reject circle with large radius by imposing a center
	// inside the shape
	if bbox := sh.BoundingBox(); !bbox.contains(circle.Center) {
		errCircle = inf
	}

	// give priority to segment for "almost" linear shapes
	if errSegment < 1 { // err is average for
		return segment
	} else if errSegment > errBezier {
		if (errSegment-errBezier)/errSegment < 0.05 { // 5%
			return segment
		}
	}

	if errSegment <= errCircle && errSegment <= errBezier {
		return segment
	} else if errBezier <= errSegment && errBezier <= errCircle {
		return bezier
	} else {
		return circle
	}
}
