package symbols

// encode an affine Trans, the composition
// of a (preserving ratio) scaling and a translation
//
//	V = | s  0 | U  + | tx |
//		| 0  s |	  | ty |
type Trans struct {
	Scale       Fl
	Translation Pos
}

var Id = Trans{Scale: 1, Translation: Pos{}}

func (tr Trans) Apply(p Pos) Pos {
	return Pos{
		X: tr.Scale*p.X + tr.Translation.X,
		Y: tr.Scale*p.Y + tr.Translation.Y,
	}
}

// Scale apply the given transformation
func (b Bezier) Scale(tr Trans) Bezier {
	return Bezier{tr.Apply(b.P0), tr.Apply(b.P1), tr.Apply(b.P2), tr.Apply(b.P3)}
}

func angleDiff(a1, a2 Fl) Fl {
	if a1 >= 0 {
		if a2 < 0 {
			return Min(abs(a1-a2), abs(a1-a2-360))
		}
	} else {
		if a2 > 0 {
			return Min(abs(a1-a2), abs(a1-a2+360))
		}
	}
	return abs(a1 - a2)
}

// measure how U and V are similar
func (U Bezier) distance(V Bezier) Fl {
	// handle points
	pU, isUPoint := U.IsPoint()
	pV, isVPoint := V.IsPoint()
	if isUPoint && isVPoint {
		return distP(pU, pV)
	} else if isUPoint != isVPoint {
		return Inf
	}

	var (
		curvatureDiff     Fl
		distancePointDiff Fl
		derivativeDiff    Fl
	)
	for t := 1; t < 20; t++ {
		t := Fl(t) / 20

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

	var penalityRatio Fl = 1
	if distanceAngle := angleDiff(tU, tV); distanceAngle > 120 {
		penalityRatio += 0.5
	}
	if du, dv := U.diffWithLine(), V.diffWithLine(); du < 0.1 && dv > 0.2 || dv < 0.1 && du > 0.2 {
		penalityRatio += 0.5
	}

	distancePointDiff /= 200
	curvatureDiff *= 10 // to be comparable with the other metrics

	distanceControls := U.P0.Sub(V.P0).NormSquared() + U.P3.Sub(V.P3).NormSquared() +
		0.05*(U.P1.Sub(V.P1).NormSquared()+U.P2.Sub(V.P2).NormSquared())
	distanceControls /= 16

	// fmt.Println("Bezier distance :", derivativeDiff*10, curvatureDiff, distancePointDiff, distanceControls, penalityRatio)

	return (derivativeDiff*10 + curvatureDiff + distancePointDiff + distanceControls) * penalityRatio
}

func (sh Shape) scale(tr Trans) Shape {
	out := make(Shape, len(sh))
	for i, p := range sh {
		out[i] = tr.Apply(p)
	}
	return out
}

func shapeDistance(s1, s2 Shape) Fl {
	s1 = removeSideArtifacts(s1)
	s2 = removeSideArtifacts(s2)

	tr := mapFromTo(s1.BoundingBox(), s2.BoundingBox())
	s1 = s1.scale(tr)

	cbox := s1.BoundingBox()
	h, w := cbox.Height(), cbox.Width()
	if w < 0.1 { // handle linear sections
		w = 1
	}
	if h < 0.1 { // handle linear sections
		h = 1
	}
	tr = Trans{Scale: 10 / Sqrt(h*w)}
	s1 = s1.scale(tr)
	s2 = s2.scale(tr)

	arcLength1 := pathLengthIndices(s1)
	arcLength2 := pathLengthIndices(s2)
	// normalize 10 [0,100]
	for i := range arcLength1 {
		arcLength1[i] *= 100
	}
	for i := range arcLength2 {
		arcLength2[i] *= 100
	}

	var worstDistance Fl
	for i1, p1 := range s1 {
		// compute the minimum distance between each curves
		min := Inf
		for i2, p2 := range s2 {
			d := dist3(p1, arcLength1[i1], p2, arcLength2[i2])
			if d < min {
				min = d
			}
		}
		if min > worstDistance {
			worstDistance = min
		}
	}
	return worstDistance
}

func dist3(p1 Pos, f1 Fl, p2 Pos, f2 Fl) Fl {
	return p1.Sub(p2).NormSquared() + (f1-f2)*(f1-f2)
}
