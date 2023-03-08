package symbols

import (
	"math"
)

// [Lookup] performs approximate matching by finding
// the closest shape in the database and returning its rune.
// More precisely, it compares scores for [rec.Shape()] and [rec.Compound()]
// returning which is better in [preferCompound].
//
// It will panic is the store is empty.
//
// # Matching overview
//
// To match an input [Record] against the database,
// we perform the following steps :
//   - for each [Shape] in the record, segment it in elementary [ShapeAtom], yielding a [ShapeFootprint]
//   - for each symbol entry in the database, compute the distance between its footprint and the input
//   - TODO: if several results have the lower distance, use the one with the lower mapping determinant
//   - perform these steps for the last [Shape] and for the whole [Symbol], and keep the best
func (db *SymbolStore) Lookup(rec Record) (r rune, preferCompound bool) {
	var (
		bestIndexCompound, bestIndexLast int
		bestDistCompound, bestDistLast   fl = inf, inf
		bestTrCompound, bestTrLast       trans
	)

	compound := rec.Compound().SegmentToAtoms()
	last := Symbol{rec.Shape()}.SegmentToAtoms()

	for i, entry := range db.entries {
		distCompound, trCompound := inf, trans{}
		distLast, trLast := inf, trans{}
		// only try symbols with same number of atoms
		if len(compound) == len(entry.Shape) {
			distCompound, trCompound = distanceFootprints(compound, entry.Shape)
		}
		if len(last) == len(entry.Shape) {
			distLast, trLast = distanceFootprints(last, entry.Shape)
		}

		if distCompound < bestDistCompound {
			bestIndexCompound = i
			bestDistCompound = distCompound
			bestTrCompound = trCompound
		}
		if distLast < bestDistLast {
			bestIndexLast = i
			bestDistLast = distLast
			bestTrLast = trLast
		}
	}

	_, _ = bestTrLast.det(), bestTrCompound.det()

	// If shape is adjacent to compound, always prefer compound
	// do not normalize !
	if closestPointDistance(rec.Shape(), rec.LastCompound().Union()) < penWidth {
		return db.entries[bestIndexCompound].R, true
	}

	if bestDistCompound < bestDistLast {
		return db.entries[bestIndexCompound].R, true
	}
	return db.entries[bestIndexLast].R, false
}

// smooth applies a moving average smoothing
func (sh Shape) smooth() Shape {
	out := make(Shape, len(sh))
	for i, p := range sh {
		if i == 0 || i == len(sh)-1 {
			out[i] = p
			continue
		}
		before, after := sh[i-1], sh[i+1]
		out[i].X = (before.X + p.X + after.X) / 3.
		out[i].Y = (before.Y + p.Y + after.Y) / 3.
	}
	return out
}

// directions returns a list of n-1 angles, giving
// the direction of the pen movement with respect to
// an horizontal reference
func (sh Shape) directions() []fl {
	if len(sh) <= 1 {
		return nil
	}

	out := make([]fl, len(sh)-1)
	for i := range out {
		// compute the direction vector
		start, end := sh[i], sh[i+1]
		u := Pos{X: end.X - start.X, Y: end.Y - start.Y}

		// we want a "continuous" angle, that is we want to avoid
		// jumps coming from the principal measure of the angle
		// To do so, we compute angles as delta from the previous
		// direction

		if i != 0 { // measure the angle variation, by taking the previous direction as reference
			prev := sh[i-1]
			v := Pos{X: start.X - prev.X, Y: start.Y - prev.Y}
			delta := angle(u, v)
			out[i] = out[i-1] + delta
		} else {
			// start with the horizontal line as reference
			v := Pos{X: 1, Y: 0}
			out[i] = angle(u, v)
		}
	}
	return out
}

// return angle in degree
func angle(u, v Pos) fl {
	dot := float64(u.X*v.X + u.Y*v.Y) // u.v
	det := float64(u.X*v.Y - u.Y*v.X) // u ^ v
	angle := math.Atan2(det, dot)     // in radian
	return fl(angle * 180 / math.Pi)  // in degre
}

func minMax(values []fl) (min, max fl) {
	min, max = values[0], values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	return min, max
}

func (sh Shape) segment() (out []Shape) {
	if len(sh) < 2 {
		return Symbol{sh}
	}
	angles := sh.smooth().directions()

	// compute the sub shapes
	clusters := segmentation(angles)

	// identify each subshape
	for _, cl := range clusters {
		subShape := sh[cl[0]:cl[1]]
		out = append(out, subShape)
	}
	return out
}

// SegmentToAtoms segments the given symbol into
// simpler elementary blocks
func (sy Symbol) SegmentToAtoms() (out ShapeFootprint) {
	segments := sy.segment()
	out = make(ShapeFootprint, len(segments))
	for i, subShape := range segments {
		// identify each subshape
		out[i] = subShape.identify()
	}
	return out
}

// segment segments the given shape into
// simpler elementary blocks
func (sy Symbol) segment() (out []Shape) {
	for _, subShape := range sy {
		// segment each subshape
		out = append(out, subShape.segment()...)
	}
	return out
}
