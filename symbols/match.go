package symbols

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// normalizeTo computes the affine 2D transformation sending the shape
// bbox to [target] (typically the reference character),
// and applies it to the coordinates of the shape
func (sh Shape) normalizeTo(target Rect) Shape {
	shapeBB := sh.BoundingBox()
	if shapeBB.IsEmpty() || target.IsEmpty() {
		return sh
	}
	// we look for a affine function f(x,y) = (ax + b, cy + d)
	// such that f(shapeBB) = target
	// a := (target.LR.X - target.UL.X) / (shapeBB.LR.X - shapeBB.UL.X)
	b := target.LR.X - 1*shapeBB.LR.X

	// c := (target.LR.Y - target.UL.Y) / (shapeBB.LR.Y - shapeBB.UL.Y)
	d := target.LR.Y - 1*shapeBB.LR.Y

	out := make(Shape, len(sh))
	for i, p := range sh {
		out[i] = Pos{1*p.X + b, 1*p.Y + d}
	}
	return out
}

func (db *SymbolStore) match(rec Record) (rune, bool) {
	var (
		bestIndexCompound, bestIndexShape int
		bestDistCompound, bestDistShape   float32 = math.MaxFloat32, math.MaxFloat32
	)

	compoundShape := rec.Compound().Union()
	shape := rec.Shape()

	for i, entry := range db.entries {
		entryS := entry.Shape.smooth()
		entryBB := entryS.BoundingBox()
		// we align the input bounding box to each of the candidate
		normalizedShape := shape.normalizeTo(entryBB).smooth()
		normalizedCompoundShape := compoundShape.normalizeTo(entryBB).smooth()

		if score := frechetDistanceShapes(normalizedCompoundShape, entryS); score < bestDistCompound {
			bestIndexCompound = i
			bestDistCompound = score
		}
		if score := frechetDistanceShapes(normalizedShape, entryS); score < bestDistShape {
			bestIndexShape = i
			bestDistShape = score
		}
	}

	// If shape is adjacent to compound, always prefer compound
	// do not normalize !
	if closestPointDistance(rec.Shape(), rec.LastCompound().Union()) < penWidth {
		return db.entries[bestIndexCompound].R, true
	}

	if bestDistCompound < bestDistShape {
		return db.entries[bestIndexCompound].R, true
	}
	return db.entries[bestIndexShape].R, false
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
		out[i].X = (before.X + p.X + after.X) / 3
		out[i].Y = (before.Y + p.Y + after.Y) / 3
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

		// we cant a "continuous" angle, that is we want to avoid
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

// return a rough version of a graph (i, values[i])
func graph(values []fl) image.Image {
	// adjust the graph height
	min, max := minMax(values)
	out := image.NewGray(image.Rect(0, int(-max)/2, len(values), int(-min)/2))
	for i, v := range values {
		out.SetGray(i, int(-v)/2, color.Gray{255})
	}
	return out
}

// diff compute the discrete derivative, return a slice with length n-1
func diff(values []fl) []fl {
	if len(values) <= 2 {
		return nil
	}
	out := make([]fl, len(values)-1)
	for i := range out {
		out[i] = values[i+1] - values[i]
	}
	return out
}

func (sh Shape) PixelImg() image.Image {
	rect := sh.BoundingBox()
	img := image.NewGray(image.Rect(int(rect.LR.X), int(rect.LR.Y), int(rect.UL.X), int(rect.UL.Y)))
	for _, p := range sh {
		img.SetGray(int(p.X), int(p.Y), color.Gray{255})
	}
	return img
}

func (sh Shape) AngularGraph() image.Image {
	return graph(sh.smooth().directions())
	// return graph(diff(sh.smoothTo().directions()))
}

// Segment segments the given shape into
// simpler elementary blocks
func (sh Shape) Segment() (out []ShapeAtom) {
	// sh = sh.smooth()
	angles := sh.smooth().directions()

	// adjust the scale and build Pos array
	min, _ := minMax(angles)
	toSegment := make([]Pos, len(angles))
	for i, a := range angles {
		toSegment[i] = Pos{X: float32(i), Y: a - min}
	}

	// compute the sub shapes
	clusters := segmentation(toSegment)

	// identify each subshape
	for _, cl := range clusters {
		subShape := sh[cl[0]:cl[1]]
		out = append(out, subShape.identify())
	}
	return out
}

func (sh Shape) AngleClustersGraph() image.Image {
	angles := sh.smooth().directions()
	// adjust the scale
	min, _ := minMax(angles)
	toSegment := make([]Pos, len(angles))
	for i, a := range angles {
		toSegment[i] = Pos{X: float32(i), Y: a - min}
	}

	clusters := segmentation(toSegment)
	fmt.Println("K", len(clusters))

	// min, max := minMax(angles)
	// out := image.NewNRGBA(image.Rect(0, int(-max)/2, len(angles), int(-min)/2))
	// for i, v := range angles {
	// 	cl := int(clusters[i]) + 1
	// 	// color according to the cluster
	// 	out.SetNRGBA(i, int(-v)/2, color.NRGBA{uint8(cl * 255 / K), 0, 0, 255})
	// }

	return nil
}
