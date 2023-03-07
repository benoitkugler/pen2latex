package symbols

import "math"

// implementation of the discrete Fréchet distance,
// an approximation of the Fréchet metric for polygonal curves, defined by Eiter and Mannila in
// Eiter, Thomas; Mannila, Heikki (1994), Computing discrete Fréchet distance
// http://www.kr.tuwien.ac.at/staff/eiter/et-archive/cdtr9464.pdf

// Discrete Fréchet distance (positive, lower is better).
// See Eiter and Mannila (1994), Table 1: Algorithm computing the coupling measure
func frechetDistanceShapes(u, v Shape) fl {
	p, q := len(u), len(v)
	// couplings, initialized to -1
	ca := make([][]fl, p)
	for i := range ca {
		ca[i] = make([]fl, q)
		for j := range ca[i] {
			ca[i][j] = -1
		}
	}

	var computeCoupling func(i, j int) fl
	computeCoupling = func(i, j int) fl {
		if ca[i][j] > -1 {
			return ca[i][j]
		} else if i == 0 && j == 0 {
			ca[i][j] = distP(u[0], v[0])
		} else if i > 0 && j == 0 {
			ca[i][j] = max(computeCoupling(i-1, 0), distP(u[i], v[0]))
		} else if i == 0 && j > 0 {
			ca[i][j] = max(computeCoupling(0, j-1), distP(u[0], v[j]))
		} else if i > 0 && j > 0 {
			d1 := computeCoupling(i-1, j)
			d2 := computeCoupling(i-1, j-1)
			d3 := computeCoupling(i, j-1)
			d4 := distP(u[i], v[j])
			ca[i][j] = max(min(min(d1, d2), d3), d4)
		}
		return ca[i][j]
	}
	return computeCoupling(p-1, q-1)
}

// return the distance between the closest point between [u] and [v]
// This NOT a measure of similarity
func closestPointDistance(u, v Shape) fl {
	best := fl(math.Inf(1))
	for _, pu := range u {
		for _, pv := range v {
			if d := distP(pu, pv); d < best {
				best = d
			}
		}
	}
	return best
}

// encode an affine trans, the composition
// of a (preserving ratio) scaling and a translation
//
//	V = | s  0 | U  + | tx |
//		| 0  s |	  | ty |
type trans struct {
	s fl
	t Pos
}

var id = trans{s: 1, t: Pos{}}

func (tr *trans) det() fl { return tr.s * tr.s }

func (tr trans) apply(p Pos) Pos {
	return Pos{
		X: tr.s*p.X + tr.t.X,
		Y: tr.s*p.Y + tr.t.Y,
	}
}

func (seg Segment) scale(tr trans) ShapeAtom {
	return Segment{Start: tr.apply(seg.Start), End: tr.apply(seg.End)}
}

func (ci Circle) scale(tr trans) ShapeAtom {
	return Circle{Center: tr.apply(ci.Center), Radius: ci.Radius.ScaleTo(abs(tr.s))}
}

func (b Bezier) scale(tr trans) ShapeAtom {
	return Bezier{tr.apply(b.P0), tr.apply(b.P1), tr.apply(b.P2), tr.apply(b.P3)}
}

func (U Segment) distance(V Segment) fl {
	d1 := U.Start.Sub(V.Start).NormSquared() + U.End.Sub(V.End).NormSquared()
	d2 := U.Start.Sub(V.End).NormSquared() + U.End.Sub(V.Start).NormSquared()
	return min(d1, d2) / 2
}

// compute the transformation needed to map [U] to [V] (as close as possible)
func (U Segment) getMapTo(V Segment) trans {
	Lu, Lv := distP(U.Start, U.End), distP(V.Start, V.End)
	s := Lv / Lu
	// align barycenter of scaled U
	Bu := U.Start.Add(U.End).ScaleTo(0.5 * s)
	Bv := V.Start.Add(V.End).ScaleTo(0.5)

	t := Bv.Sub(Bu)

	return trans{s, t}
}

func (U Circle) distance(V Circle) fl {
	dc := U.Center.Sub(V.Center).NormSquared()
	dr := U.Radius.Sub(V.Radius).NormSquared()
	return (dc + dr) / 2
}

func (U Circle) getMapTo(V Circle) trans {
	sx := V.Radius.X / U.Radius.X
	sy := V.Radius.Y / U.Radius.Y
	s := sqrt(sx * sy)
	t := V.Center.Sub(U.Center)
	return trans{s, t}
}

func (U Bezier) getMapTo(V Bezier) trans {
	// only use start and end to compute the transformation
	return Segment{U.P0, U.P3}.getMapTo(Segment{V.P0, V.P3})
}

func (U Bezier) distance(V Bezier) fl {
	d1 := (U.P0.Sub(V.P0).NormSquared() +
		U.P1.Sub(V.P1).NormSquared() +
		U.P2.Sub(V.P2).NormSquared() +
		U.P3.Sub(V.P3).NormSquared())
	d2 := (U.P0.Sub(V.P3).NormSquared() +
		U.P1.Sub(V.P2).NormSquared() +
		U.P2.Sub(V.P1).NormSquared() +
		U.P3.Sub(V.P0).NormSquared())
	return min(d1, d2) / 4
}

func (U Bezier) distanceCircle(V Circle) fl {
	// discretize U and average
	var d fl
	for i := 0; i <= 10; i++ {
		t := fl(i) / 10
		d += V.squaredDistancePoint(U.eval(t))
	}
	return d / (10 + 1)
}

// return a mapping from U to V
func mapBetweenAtoms(U, V ShapeAtom) (trans, bool) {
	switch U := U.(type) {
	case Segment:
		V, ok := V.(Segment)
		if !ok {
			return trans{}, false
		}
		return U.getMapTo(V), true
	case Bezier:
		V, ok := V.(Bezier)
		if !ok {
			return trans{}, false
		}
		return U.getMapTo(V), true
	case Circle:
		V, ok := V.(Circle)
		if !ok {
			return trans{}, false
		}
		return U.getMapTo(V), true
	default:
		panic("exhaustive type switch")
	}
}

func distanceBetweenAtoms(U, V ShapeAtom) fl {
	switch U := U.(type) {
	case Segment:
		V, ok := V.(Segment)
		if !ok {
			return inf
		}
		return U.distance(V)
	case Bezier:
		switch V := V.(type) {
		case Bezier:
			return U.distance(V)
		case Circle:
			return U.distanceCircle(V)
		default:
			return inf
		}
	case Circle:
		switch V := V.(type) {
		case Circle:
			return U.distance(V)
		case Bezier:
			return V.distanceCircle(U)
		default:
			return inf
		}
	default:
		panic("exhaustive type switch")
	}
}

// compute the distance between two lists which must have the same length,
// by computing the optimal permutation and transformation from one to the other
func distanceFootprints(U, V []ShapeAtom) (fl, trans) {
	best := inf
	var bestTr trans
	perm(U, func(permuted []ShapeAtom) {
		d, tr := permShapeDistance(permuted, V)
		if d < best {
			best = d
			bestTr = tr
		}
	}, 0)
	return best, bestTr
}

func permShapeDistanceScaled(Us, Vs []ShapeAtom, tr trans) fl {
	var totalDistance fl
	for j, U := range Us {
		U = U.scale(tr)
		totalDistance += distanceBetweenAtoms(U, Vs[j])
	}

	return totalDistance / fl(len(Us)) // normalize
}

func permShapeDistance(Us, Vs []ShapeAtom) (fl, trans) {
	best := inf
	var bestTr trans
	for i := range Us {
		tr, ok := mapBetweenAtoms(Us[i], Vs[i])
		if !ok {
			continue
		}
		if d := permShapeDistanceScaled(Us, Vs, tr); d < best {
			best = d
			bestTr = tr
		}
	}
	return best, bestTr
}

// Permute the values at index i to len(a)-1.
// mutating a
func perm(a []ShapeAtom, f func([]ShapeAtom), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}
