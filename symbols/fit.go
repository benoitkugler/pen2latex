package symbols

import (
	"fmt"
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
func fitCubicBezier(points []Pos) (Bezier, fl) {
	const maxIterations = 8 // tuned experimentally

	// unit tangent vectors at endpoints
	tHat1 := computeLeftTangent(points, 0)
	tHat2 := computeRightTangent(points, len(points)-1)

	bestErr, bestBezier := inf, Bezier{}

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
		totalLength += distP(points[i], points[i-1])
		out[i] = totalLength
	}
	// normalize to [0,1]
	for i := range out {
		out[i] /= totalLength
	}
	return out
}

// u is the current parametrization of the points
func generateBezier(d []Pos, u []fl, tHat1, tHat2 Pos) Bezier {
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

		bez := Bezier{d[first], d[first], d[last], d[last]}.eval(u[i])

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
	segLength := distP(d[first], d[last])
	epsilon := 1.0e-6 * segLength
	if alpha_l < epsilon || alpha_r < epsilon {
		/* fall back on standard (probably inaccurate) formula, and subdivide further if needed. */
		dist := segLength / 3.0
		return Bezier{
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
	return Bezier{
		P0: d[first],
		P1: tHat1.ScaleTo(alpha_l).Add(d[first]),
		P2: tHat2.ScaleTo(alpha_r).Add(d[last]),
		P3: d[last],
	}
}

// Given set of points and their parameterization, try to find
// a better parameterization.
func reparameterize(d []Pos, u []fl, bezCurve Bezier) []fl {
	uPrime := make([]fl, len(d)) //  new parameter values
	for i := range d {
		uPrime[i] = newtonRaphsonRootStep(bezCurve, d[i], u[i])
	}
	return uPrime
}

// newtonRaphsonRootStep :
// Use Newton-Raphson iteration to find better root.
func newtonRaphsonRootStep(Q Bezier, P Pos, u fl) fl {
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
func computeBezierDistance(d []Pos, bezCurve Bezier, u []fl) fl {
	var average fl // maximum fitError
	for i := range d {
		P := bezCurve.eval(u[i]) // pos on curve
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
	Radius Pos // rx, ry
}

// compute the barycentre of the points
func computeBarycentre(points []Pos) Pos {
	var barycentre Pos
	for _, p := range points {
		barycentre.X += p.X
		barycentre.Y += p.Y
	}
	barycentre.Scale(1 / fl(len(points)))
	return barycentre
}

func fitCircle(points []Pos) (Circle, fl) {
	var tx2, ty2, txy fl      // x^2, y^2, xy
	var s30, s21, s12, s03 fl // x^3, x^2y, xy^2, y^3

	// compute barycentre
	barycentre := computeBarycentre(points)

	for _, p := range points {
		// shift by the barycentre
		x, y := p.X-barycentre.X, p.Y-barycentre.Y
		x2, y2 := x*x, y*y

		tx2 += x2
		ty2 += y2
		txy += x * y

		s30 += x * x2
		s21 += x2 * y
		s12 += x * y2
		s03 += y2 * y
	}
	denom := 2 * (tx2*ty2 - txy*txy)
	cx := ((s12+s30)*ty2 - (s21+s03)*txy) / denom
	cy := ((s21+s03)*tx2 - (s12+s30)*txy) / denom
	center := Pos{cx, cy}
	N := fl(len(points))
	r := sqrt(cx*cx + cy*cy + (tx2+ty2)/N)

	// translate by the barycentre
	center = center.Add(barycentre)

	// error value
	var averageErr fl
	for _, p := range points {
		tmp := p.Sub(center).Norm() - r // maybe be negative
		averageErr += tmp * tmp
	}
	return Circle{center, Pos{r, r}}, averageErr / N
}

func fitEllipse(points []Pos) (Circle, fl) {
	// adapted from the circle case
	// given the equation
	// ((x - c_x)/a)² + ((y - c_y)/b)² = 1
	// we use the variables R = a and S = a/b
	// yielding the following relation (closer to the circle)
	// (x - c_x)² + S² (y - c_y) = R²

	var tx2, ty2, txy fl      // x^2, y^2, xy
	var s30, s21, s12, s03 fl // x^3, x^2y, xy^2, y^3
	var y4, x2y2 fl           // y^4, x2*y2

	// compute barycentre
	barycentre := computeBarycentre(points)

	for _, p := range points {
		// shift by the barycentre
		x, y := p.X-barycentre.X, p.Y-barycentre.Y
		x2, y2 := x*x, y*y

		tx2 += x2
		ty2 += y2
		txy += x * y

		s30 += x * x2
		s21 += x2 * y
		s12 += x * y2
		s03 += y2 * y

		y4 += y2 * y2
		x2y2 += x2 * y2
	}

	N := fl(len(points))

	// solve the system link c_x, c_y, S
	a, b, d, e := 2*tx2, txy, -s12, s30
	a_, b_, d_, e_ := 2*txy, ty2, -s03, s21
	a__, b__, d__, e__ := -2*s12, -2*s03, -(y4 - (ty2 * ty2 / N)), x2y2-(tx2*ty2/N)

	alpha, beta, gamma := a_*b-a*b_, a_*d-a*d_, a_*e-a*e_
	alpha_, beta_, gamma_ := a__*b-a*b__, a__*d-a*d__, a__*e-a*e__

	S2 := (alpha_*gamma - gamma_*alpha) / (alpha_*beta - alpha*beta_)
	cy := ((gamma / S2) - beta) / alpha
	cx := (e - S2*(b*cy+d)) / a
	R2 := cx*cx + tx2/N + S2*(cy*cy+ty2/N)

	center := Pos{cx, cy}

	ra := sqrt(R2)
	rb := ra / sqrt(S2)

	// translate by the barycentre
	center = center.Add(barycentre)

	// error value :
	// we compute the distance to the ellipse by
	// transforming it to a C(0, 1) circle
	// and applying the scaling back
	var averageErr fl
	for _, p := range points {
		// translate to the 0
		p = p.Sub(center)
		// apply the linear function (x,y) ->(x/a, y/b)
		p.X /= ra
		p.Y /= rb
		// compute the vector between the projection on the C(0,1)
		// and p
		proj := p.ScaleTo(1 / p.Norm())
		diff := p.Sub(proj)
		// apply the inverse linear transformation (x,y) -> (ax, by)
		dist := (ra*diff.X)*(ra*diff.X) + (rb*diff.Y)*(rb*diff.Y)
		averageErr += dist
	}

	return Circle{center, Pos{ra, rb}}, averageErr / N
}

func nanToInf(v fl) fl {
	if math.IsNaN(float64(v)) {
		return inf
	}
	return v
}

func fitCircleOrEllipse(points []Pos) (Circle, fl) {
	// try the general ellipse
	ellipse, errE := fitEllipse(points)
	circle, errC := fitCircle(points)
	errE, errC = nanToInf(errE), nanToInf(errC)

	if errC <= errE {
		return circle, errC
	}
	return ellipse, errE
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
	denom := (N*sx2 - sx*sx)

	var u, A Pos           // vector director and point in the line
	if abs(denom) < 1e-6 { // handle vertical line
		u = Pos{0, 1}
		A = Pos{points[0].X, 0}
	} else {
		m := (N*sxy - sx*sy) / denom
		b := (sy - m*sx) / N
		// vector director
		u = Pos{1, m}
		A = Pos{0, b}
	}
	nu := u.NormSquared()

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

func (k ShapeAtomKind) String() string {
	switch k {
	case SAKBezier:
		return "Bezier"
	case SAKSegment:
		return "Segment"
	case SAKCircle:
		return "Circle"
	default:
		panic("invalid kind")
	}
}

type shapeAtomData struct {
	Data []Pos         `json:"d"`
	Kind ShapeAtomKind `json:"k"`
}

func (d shapeAtomData) deserialize() (ShapeAtom, error) {
	switch d.Kind {
	case SAKBezier:
		if L := len(d.Data); L != 4 {
			return nil, fmt.Errorf("invalid length for BezierC %d", L)
		}
		return Bezier{d.Data[0], d.Data[1], d.Data[2], d.Data[3]}, nil
	case SAKSegment:
		if L := len(d.Data); L != 2 {
			return nil, fmt.Errorf("invalid length for Segment %d", L)
		}
		return Segment{d.Data[0], d.Data[1]}, nil
	case SAKCircle:
		if L := len(d.Data); L != 2 {
			return nil, fmt.Errorf("invalid length for Circle %d", L)
		}
		return Circle{Center: d.Data[0], Radius: d.Data[1]}, nil
	default:
		return nil, fmt.Errorf("invalid ShapeAtomKind %d", d.Kind)
	}
}

func (b Bezier) serialize() shapeAtomData {
	return shapeAtomData{Kind: SAKBezier, Data: []Pos{b.P0, b.P1, b.P2, b.P3}}
}

func (s Segment) serialize() shapeAtomData {
	return shapeAtomData{Kind: SAKSegment, Data: []Pos{s.Start, s.End}}
}

func (c Circle) serialize() shapeAtomData {
	return shapeAtomData{Kind: SAKCircle, Data: []Pos{c.Center, c.Radius}}
}

type ShapeAtom interface {
	Kind() ShapeAtomKind

	scale(trans) ShapeAtom // with same concrete type

	serialize() shapeAtomData
}

func (Bezier) Kind() ShapeAtomKind  { return SAKBezier }
func (Segment) Kind() ShapeAtomKind { return SAKSegment }
func (Circle) Kind() ShapeAtomKind  { return SAKCircle }

func (sh Shape) identify() ShapeAtom {
	if len(sh) == 0 {
		return Bezier{}
	} else if len(sh) <= 2 {
		start, end := sh[0], sh[len(sh)-1]
		return Segment{start, end}
	}

	bezier, errBezier := fitCubicBezier(sh)
	segment, errSegment := fitSegment(sh)
	circle, errCircle := fitCircleOrEllipse(sh)

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
