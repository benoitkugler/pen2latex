package symbols

import "fmt"

// This file implements algorithms to fit curves to points :
//	- for a line
//	- for a cubic bezier curve
//	- for an union of bezier curves

// --------------------------------------------------------------------
// ----------------------------- Line fit -----------------------------
// --------------------------------------------------------------------

type segment struct{ p0, p1 Pos }

func (s segment) asBezier() Bezier {
	p1 := s.p0.Add(s.p1).ScaleTo(0.5)
	return Bezier{s.p0, p1, p1, s.p1}
}

// use a least square linear regression
func fitSegment(points []Pos) (segment, Fl) {
	// find the line ...
	var sx, sy, sxy, sx2 Fl
	N := Fl(len(points)) - 2
	for _, p := range points[1 : len(points)-1] {
		x, y := (p.X), (p.Y)
		sx += x
		sy += y
		sxy += x * y
		sx2 += x * x
	}
	sx /= N
	sy /= N
	sxy /= N
	sx2 /= N

	// y = mx + b
	denom := (sx2 - sx*sx)

	var u, A Pos           // vector director and point in the line
	if abs(denom) < 1e-4 { // handle vertical line
		u = Pos{0, 1}
		A = Pos{points[0].X, 0}
	} else {
		m := (sxy - sx*sy) / denom
		b := (sy - m*sx)
		// vector director
		u = Pos{1, Fl(m)}
		A = Pos{0, Fl(b)}
	}
	nu := u.NormSquared()

	// .. then find start and end
	minT, maxT := Inf, -Inf
	var (
		start, end Pos
		err        Fl
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

	return segment{start, end}, err / N
}

// ------------------------------------------------------------------
// ------------------------ One cubic bezier fit --------------------
// ------------------------------------------------------------------

// fitCubicBezier implements a gradient descent for fitting ONE cubic bezier curve
// to the given points, returning the average quadratic error computed by [bezierError]
//
// if panics if len(d) < 2
func fitCubicBezier(points []Pos) (Bezier, Fl) {
	const maxIterations = 50

	pathLengths := pathLengthIndices(points)
	start, end := points[0], points[len(points)-1]
	bezier := Bezier{P0: start, P3: end}
	// start with aligned control points
	bezier.P1 = start.ScaleTo(2).Add(end).ScaleTo(1. / 3.)
	bezier.P2 = start.Add(end.ScaleTo(2)).ScaleTo(1. / 3.)

	errValue := Inf
	for i := 0; i < maxIterations; i++ {
		grad := bezierEnergyGradient(points, pathLengths, bezier)
		const step = -0.1

		bezier.P1.X += step * grad[0]
		bezier.P1.Y += step * grad[1]
		bezier.P2.X += step * grad[2]
		bezier.P2.Y += step * grad[3]

		newErrValue, _ := bezierError(points, bezier, pathLengths)
		if abs(newErrValue-errValue) < 0.1 {
			break
		}
		errValue = newErrValue
	}

	return bezier, errValue
}

// return the derivatives of B(t) with respect to
// control points P1 and P2
func controlDerivatives(t Fl) (dx1, dy1, dx2, dy2 Pos) {
	s := 1 - t
	b1 := 3 * s * s * t
	dx1 = Pos{b1, 0}
	dy1 = Pos{0, b1}
	b2 := 3 * s * t * t
	dx2 = Pos{b2, 0}
	dy2 = Pos{0, b2}
	return
}

// bezierEnergyGradient returns the gradient of the distance between the points
// and the [bezier] curve, with respect to the control points
func bezierEnergyGradient(points []Pos, pathLengths []Fl, bezier Bezier) (out [4]Fl) {
	for i, ti := range pathLengths {
		bti := bezier.pointAt(ti)
		ui := bti.Sub(points[i])
		dx1, dy1, dx2, dy2 := controlDerivatives(ti)
		out[0] += dotProduct(ui, dx1)
		out[1] += dotProduct(ui, dy1)
		out[2] += dotProduct(ui, dx2)
		out[3] += dotProduct(ui, dy2)
	}
	return out
}

// ------------------------------------------------------------------------------------
// ---------------------------- Union of bezier curves fit ----------------------------
// ------------------------------------------------------------------------------------

// fitCubicBeziers implements the algorithm proposed in
//
// An Algorithm for Automatically Fitting Digitized Curves
// by Philip J. Schneider
// from "Graphics Gems", Academic Press, 1990
//
// and popularized in
// https://stackoverflow.com/questions/5525665/smoothing-a-hand-drawn-curve/5530600#5530600
func fitCubicBeziers(points []Pos) []Bezier {
	// remove artifacts
	points = removeSideArtifacts(points)

	if len(points) == 1 {
		p := points[0]
		return []Bezier{{p, p, p, p}}
	}

	// unit tangent vectors at endpoints
	tHat1 := computeStartTangent(points)
	tHat2 := computeEndTangent(points)

	return fitOrSplitCubicBeziers(points, tHat1, tHat2)
}

// points is a subslice of the original shape
func fitOrSplitCubicBeziers(points []Pos, tHat1, tHat2 Pos) []Bezier {
	const threshold = 5
	const iterationThreshold = 50

	const maxIterations = 8 // tuned experimentally

	// use heuristic if region only has two points in it
	if len(points) <= 2 {
		first, last := points[0], points[len(points)-1]
		dist := first.Sub(last).NormSquared() / 3.0
		p1, p2 := tHat1.ScaleTo(dist).Add(first), tHat2.ScaleTo(dist).Add(last)
		return []Bezier{{first, p1, p2, last}}
	}

	// start parametrization with path arc length
	u := pathLengthIndices(points)
	bezCurve := inferBezier(points, u, tHat1, tHat2)
	currentError, splitPoint := bezierError(points, bezCurve, u)

	// if the error is not too large, refine with reparameterization
	if currentError < iterationThreshold {
		bestError := currentError
		bestBezier := bezCurve
		for i := 0; i < maxIterations; i++ {
			u = reparameterize(points, u, bezCurve)
			bezCurve := inferBezier(points, u, tHat1, tHat2)
			err, split := bezierError(points, bezCurve, u)
			if err < bestError {
				bestError = err
				bestBezier = bezCurve
				splitPoint = split
			}
		}

		if bestError < threshold { // the whole shape is a bezier curve, return early
			return []Bezier{bestBezier}
		}
	}

	// fitting failed: split at max error point and fit recursively
	tHatCenter1, tHatCenter2 := computeCenterTangent(points, splitPoint)

	l1 := fitOrSplitCubicBeziers(points[:splitPoint+1], tHat1, tHatCenter1)
	l2 := fitOrSplitCubicBeziers(points[splitPoint:], tHatCenter2, tHat2)

	return append(l1, l2...)
}

// u is the current parametrization of the points
func inferBezier(d []Pos, u []Fl, tHat1, tHat2 Pos) Bezier {
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

	// create the C and X matrices
	var (
		C [2][2]Fl
		X [2]Fl
	)

	for i := range d {
		C[0][0] += dotProduct(A[i][0], A[i][0])
		C[0][1] += dotProduct(A[i][0], A[i][1])
		C[1][0] = C[0][1]
		C[1][1] += dotProduct(A[i][1], A[i][1])

		bez := Bezier{d[first], d[first], d[last], d[last]}.pointAt(u[i])

		tmp := d[first+i].Sub(bez)

		X[0] += dotProduct(A[i][0], tmp)
		X[1] += dotProduct(A[i][1], tmp)
	}

	// compute the determinants of C and X
	det_C0_C1 := C[0][0]*C[1][1] - C[1][0]*C[0][1]
	det_C0_X := C[0][0]*X[1] - C[1][0]*X[0]
	det_X_C1 := X[0]*C[1][1] - X[1]*C[0][1]

	// finally, derive alpha values
	var alpha_l, alpha_r Fl
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
func reparameterize(d []Pos, u []Fl, bezCurve Bezier) []Fl {
	uPrime := make([]Fl, len(d)) //  new parameter values
	for i := range d {
		uPrime[i] = newtonRaphsonRootStep(bezCurve, d[i], u[i])
	}
	return uPrime
}

// newtonRaphsonRootStep :
// Use Newton-Raphson iteration to find better root.
func newtonRaphsonRootStep(Q Bezier, P Pos, u Fl) Fl {
	// Compute Q(u)
	Q_u := Q.pointAt(u)

	// Compute Q'(u) and Q''(u)
	Q1_u := Q.derivativeAt(u)
	Q2_u := Q.secondDerivativeAt(u)

	// Compute f(u)/f'(u)
	numerator := (Q_u.X-P.X)*(Q1_u.X) + (Q_u.Y-P.Y)*(Q1_u.Y)
	denominator := (Q1_u.X)*(Q1_u.X) + (Q1_u.Y)*(Q1_u.Y) +
		(Q_u.X-P.X)*(Q2_u.X) + (Q_u.Y-P.Y)*(Q2_u.Y)
	if denominator == 0.0 {
		return u
	}

	// u = u - f(u)/f'(u)
	uPrime := u - (numerator / denominator)
	return uPrime
}

func hasSpuriousRepetitionStart(points []Pos) (int, bool) {
	seen := make(map[Pos]int)
	L := 10
	if len(points) < L {
		L = len(points)
	}

	for i := 0; i < L; i++ {
		p := points[i]
		if indexSeen, ok := seen[p]; ok { // should not happen in correct inputs
			return indexSeen, true
		}
		seen[p] = i
	}
	return 0, false
}

func hasSpuriousRepetitionEnd(points []Pos) (int, bool) {
	seen := make(map[Pos]bool)
	L := len(points)

	start := L - 6
	if start < 0 {
		start = 0
	}

	for i := start; i < L; i++ {
		p := points[i]
		if seen[p] { // should not happend
			return i, true
		}
		seen[p] = true
	}
	return 0, false
}

// some profiles captured with a tablet have repetitions
func removeDuplicates(points []Pos) (out []Pos) {
	last := Pos{Inf, Inf} // can't be in points

	for _, p := range points {
		if p == last { // duplicate, ignore it
			continue
		}
		out = append(out, p)
		last = p
	}
	return out
}

func removeSideArtifacts(points []Pos) []Pos {
	points = removeDuplicates(points)

	// detect non moving, non significative points
	if cut, hasRepetition := hasSpuriousRepetitionStart(points); hasRepetition {
		points = points[cut:]
	}

	if len(points) <= 3 {
		return points
	}

	if distP(points[0], points[3]) <= 1 {
		points = points[1:]
	}
	if points[2] == points[0] {
		plus2, plus1 := points[2], points[1]
		replacement := plus1.Add(plus1.Sub(plus2))
		points[0] = replacement
	}

	// detect non moving, non significative points
	if len(points) >= 5 && diameter(points[len(points)-5:]) <= 3 {
		points = points[:len(points)-3]
	} else if cut, hasRepetion := hasSpuriousRepetitionEnd(points); hasRepetion {
		points = points[:cut]
	}

	// smooth edges
	if len(points) >= 5 {
		L := len(points) - 1
		for i := 4; i >= 1; i-- {
			points[L-i] = points[L-i-1].Add(points[L-i+1]).ScaleTo(0.5)
		}
	}

	return points
}

// return the max distance
func diameter(points []Pos) Fl {
	var maxDistance Fl
	for _, p := range points {
		for _, q := range points {
			if d := p.Sub(q).NormSquared(); d > maxDistance {
				maxDistance = d
			}
		}
	}
	return Sqrt(maxDistance)
}

// computeStartTangent, ComputeRightTangent, ComputeCenterTangent :
// Approximate unit tangents at endpoints and "center" of digitized curve
func computeStartTangent(d []Pos) Pos {
	tHat1 := d[1].Sub(d[0])

	// average several tangents for a more robust result
	tHat2, tHat3 := tHat1, tHat1
	if len(d) >= 4 {
		tHat2 = d[2].Sub(d[0])
		tHat3 = d[3].Sub(d[0])
	}

	return robustTangent(tHat1, tHat2, tHat3)
}

func computeEndTangent(d []Pos) Pos {
	end := len(d) - 1
	tHat1 := d[end-1].Sub(d[end])

	// average several tangents for a more robust result
	tHat2, tHat3 := tHat1, tHat1
	if len(d) >= 4 {
		tHat2 = d[end-2].Sub(d[end])
		tHat3 = d[end-3].Sub(d[end])
	}
	return robustTangent(tHat1, tHat2, tHat3)
}

func robustTangent(tHat1, tHat2, tHat3 Pos) Pos {
	// average several tangents for a more robust result
	if abs(angle(tHat1, tHat3)) > 30 { // large variation -> average
		tHat1 = tHat1.Add(tHat2).Add(tHat3).ScaleTo(1. / 3.)
	}
	tHat1.normalize()
	return tHat1
}

// are we at the cross between two lines ?
func havePointsSharpAngle(d []Pos, center int) (bool, Fl) {
	before, after := center-10, center+10
	if before < 0 || after >= len(d) {
		return false, 0
	}
	p1, p2 := d[before:center+1], d[center:after]
	// check if the two parts of the curve are linear
	_, err1 := fitSegment(p1)
	_, err2 := fitSegment(p2)
	first, middle, last := d[before], d[center], d[after]
	a := angle(middle.Sub(first), last.Sub(middle))
	return err1 < 0.7 && err2 < 0.7, a
}

func computeCenterTangent(d []Pos, center int) (left, right Pos) {
	left = computeEndTangent(d[:center+1])
	right = computeStartTangent(d[center:])

	// avoid spurious angular variation by using more points
	before, after := center-2, center+2
	if before >= 0 && after < len(d) {
		u := d[center].Sub(d[before])
		v := d[after].Sub(d[center])

		a := angle(u, v)
		// if ok, correctedAngle := havePointsSharpAngle(d, center); ok {
		// 	a = correctedAngle
		// }
		smooth := abs(a) < 45
		if smooth { // use the average
			tHatMean := u.Add(v).ScaleTo(0.5)

			left = tHatMean.ScaleTo(-1)
			right = tHatMean
			left.normalize()
			right.normalize()
		}
	}

	return
}

// ----------------------------- Post-processing -----------------------------

// mergeSimilarCurves post process a Bezier fit to
// merge adjacent lines and curves which where split during the fit
func mergeSimilarCurves(curves []Bezier) (out []Bezier) {
	// make sure that points are properly recognized
	if len(curves) == 1 {
		if point, ok := curves[0].isAlmostPoint(); ok {
			return []Bezier{{point, point, point, point}}
		}
	}

	// remove inplace repetitions, this has to be done before segment merging
	if L := len(curves); L > 1 {
		c1, c2 := curves[0], curves[1]
		if c1.IsRoughlyLinear() && c2.IsRoughlyLinear() && areLinesMerged(c1, c2) {
			// we got a repetition : only keep the second segment

			if debugMode {
				fmt.Println("mergeSimilarCurves: removing linear repetition")
			}

			curves = curves[1:]
		}
	}

	// start with the first curve
	out = []Bezier{curves[0]}

	for i := 1; i < len(curves); i++ {
		prevCurve := out[len(out)-1] // use the last curve added to the output
		currentCurve := curves[i]
		points := append(prevCurve.toPoints(), currentCurve.toPoints()...)

		angle := tangentAngle(prevCurve, currentCurve)

		_, errSegment := fitSegment(points)
		mergedCurve, errCurve := fitCubicBezier(points)
		// check that the merging do not introduce undesired angles
		isMergedTangentCompatible := true
		if len(out) >= 2 {
			previous := out[len(out)-2]
			wasSmooth := tangentAngle(previous, prevCurve) < 10
			isMergedTangentCompatible = isMergedTangentCompatible && (!wasSmooth || tangentAngle(previous, mergedCurve) < 20)
		}
		if i+1 < len(curves) {
			next := curves[i+1]
			wasSmooth := tangentAngle(currentCurve, next) < 10
			isMergedTangentCompatible = isMergedTangentCompatible && (!wasSmooth || tangentAngle(mergedCurve, curves[i+1]) < 20)
		}

		f1, f2, spuriousCurvature := areBeziersSpuriousCurvature(prevCurve, currentCurve)

		if isAligned := angle < 10; isAligned && errSegment < 1.1 {
			// replace the last element of out
			out[len(out)-1] = segment{prevCurve.P0, currentCurve.P3}.asBezier()

			if debugMode {
				fmt.Printf("mergeSimilarCurves: %d -> merging 2 to 1 segment\n", i)
			}

		} else if isAligned := angle < 5; isAligned && errCurve < 12 && isMergedTangentCompatible {
			// replace the last element of out
			out[len(out)-1] = mergedCurve

			if debugMode {
				fmt.Printf("mergeSimilarCurves: %d -> merging 2 to 1 curve\n", i)
			}

		} else if spuriousCurvature {
			// replace the last with the correction...
			out[len(out)-1] = f1
			// and add the current with the correction
			out = append(out, f2)
		} else {
			// no merging : add the last element
			out = append(out, currentCurve)
		}
	}

	// remove spurious curvature at start and end
	if replacement, ok := out[0].hasSpuriousCurvature(); ok {
		out[0] = replacement
	}
	if replacement, ok := out[len(out)-1].hasSpuriousCurvature(); ok {
		out[len(out)-1] = replacement
	}

	return
}

// return true if there is a spurious curvature at the end of c1 or at the start of c2,
// coming from a split at the wrong place
func areBeziersSpuriousCurvature(c1, c2 Bezier) (fixed1, fixed2 Bezier, ok bool) {
	if isAligned := tangentAngle(c1, c2) < 5; !isAligned {
		return
	}

	if t, ok := c1.hasRoughEndAngle(); ok {
		// lets split
		fixed1, _ = c1.splitAt(t)
		fixed2 = Bezier{fixed1.P3, c2.P1, c2.P2, c2.P3}
		return fixed1, fixed2, true
	}

	// TODO: handle the same for c2 at start
	return
}

// ----------------------------- shared utils -----------------------------

// return t_i : coefficients in [0, 1], computed from the path length
func pathLengthIndices(points []Pos) []Fl {
	out := make([]Fl, len(points))
	var totalLength Fl
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

// bezierError finds the maximum squared distance of digitized points
// to the fitted curve, given a parametrization
// it also returns the index of the worst point
func bezierError(d []Pos, bezCurve Bezier, u []Fl) (Fl, int) {
	var (
		maxDist  Fl
		maxIndex int
	)
	for i := range d {
		P := bezCurve.pointAt(u[i]) // pos on curve
		dist := P.Sub(d[i]).NormSquared()
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}
	return maxDist, maxIndex
}
