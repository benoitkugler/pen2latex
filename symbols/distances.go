package symbols

import (
	"math"
)

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

func (tr trans) apply(p Pos) Pos {
	return Pos{
		X: tr.s*p.X + tr.t.X,
		Y: tr.s*p.Y + tr.t.Y,
	}
}

func (b Bezier) scale(tr trans) Bezier {
	return Bezier{tr.apply(b.P0), tr.apply(b.P1), tr.apply(b.P2), tr.apply(b.P3)}
}

func angleDiff(a1, a2 fl) fl {
	if a1 >= 0 {
		if a2 < 0 {
			return min(abs(a1-a2), abs(a1-a2-360))
		}
	} else {
		if a2 > 0 {
			return min(abs(a1-a2), abs(a1-a2+360))
		}
	}
	return abs(a1 - a2)
}

// measure how U and V are similar
func (U Bezier) distance(V Bezier) fl {
	var (
		curvatureDiff     fl
		distancePointDiff fl
		derivativeDiff    fl
	)
	for t := 1; t < 20; t++ {
		t := fl(t) / 20

		dd := U.pointAt(t).Sub(V.pointAt(t)).NormSquared()
		cd := abs(U.curvatureAt(t) - V.curvatureAt(t))
		du, dv := U.derivativeAt(t), V.derivativeAt(t)
		du.normalize()
		dv.normalize()
		ddd := du.Sub(dv).NormSquared()

		distancePointDiff += dd
		curvatureDiff += cd
		derivativeDiff += ddd

	}
	tU := angle(U.derivativeAt(0), U.derivativeAt(1))
	tV := angle(V.derivativeAt(0), V.derivativeAt(1))

	if distanceAngle := angleDiff(tU, tV); distanceAngle > 120 {
		return inf
	}

	distancePointDiff /= 200
	curvatureDiff *= 10 // to be comparable with the other metrics

	distanceControls := U.P0.Sub(V.P0).NormSquared() + U.P3.Sub(V.P3).NormSquared() +
		0.05*(U.P1.Sub(V.P1).NormSquared()+U.P2.Sub(V.P2).NormSquared())
	distanceControls /= 16

	return derivativeDiff*10 + curvatureDiff + distancePointDiff + distanceControls
}
