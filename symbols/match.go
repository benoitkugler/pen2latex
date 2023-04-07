package symbols

import (
	"fmt"
	"strings"
)

const debugMode = true

// [Lookup] performs approximate matching by finding
// the closest symbol to [input] in the database and returning its rune.
//
// It will return 0 if the store is empty or if no symbol matches [input].
// It returns the best error found.
// It also returns a bool indicating if the input is compatible with an other
// input, in the sense that the first strokes are close.
//
// # Matching overview
//
// To match an input [Record] against the database,
// we perform the following steps :
//   - for each [Shape] in the record, segment it into Bezier curves, yielding a [ShapeFootprint]
//   - for each symbol entry in the database, compute the distance between its footprint and the input
//   - disambiguate results using the size of the surrounding context
func (db *Store) Lookup(input Footprint, context Context) (rune, Fl, bool) {
	var (
		bestIndex                            int = -1
		bestDistance, bestDistanceCompatible     = Inf, Inf
	)

	for i, entry := range db.Symbols {
		distExact := distanceSymbolsExact(entry.Footprint, input)
		if distExact < bestDistance {
			bestDistance = distExact
			bestIndex = i
		}
		// inspect the symbols in the store with more strokes
		if len(entry.Footprint.Strokes) > len(input.Strokes) {
			if d := distanceSymbolsCompatible(entry.Footprint, input); d < bestDistanceCompatible {
				bestDistanceCompatible = d
			}
		}
	}

	fmt.Println(bestDistanceCompatible, 2*bestDistance)
	hasCompatible := bestDistanceCompatible < Inf && bestDistanceCompatible < 2*bestDistance

	if bestIndex == -1 {
		return 0, Inf, hasCompatible
	}

	r, d := db.Symbols[bestIndex].R, bestDistance
	r = distinguishByContext(input, context, r)

	return r, d, hasCompatible
}

// ---------------------------------------------------------------------------

// Footprint builds the footprint of the symbol
func (sy Symbol) Footprint() Footprint { return newSymbolFootprint(sy) }

// Stroke stores a simplified representation of one
// [Shape]
type Stroke struct {
	Curves     []Bezier `json:"c"`
	ArcLengths []Fl     `json:"a"` // between 0 and 1, starts after the first part and ends at 1
}

func newFp(points Shape) Stroke {
	// fit and regularize
	curves := mergeSimilarCurves(fitCubicBeziers(points))
	out := Stroke{Curves: curves}

	// compute arc lengths
	out.inferArcLengths()

	return out
}

func (fp *Stroke) inferArcLengths() {
	arcLengths := make([]Fl, len(fp.Curves))
	var totalLength Fl
	for i, part := range fp.Curves {
		L := part.arcLength()
		totalLength += L
		arcLengths[i] = totalLength
	}

	// normalize
	for i := range arcLengths {
		arcLengths[i] /= totalLength
	}
	fp.ArcLengths = arcLengths
}

// return the same curve, but draw in reverse order
func (fp Stroke) reverse() Stroke {
	Lc := len(fp.Curves)
	out := Stroke{
		Curves: make([]Bezier, Lc),
	}
	// reverse curves
	for i, c := range fp.Curves {
		out.Curves[Lc-1-i] = Bezier{c.P3, c.P2, c.P1, c.P0}
	}
	out.inferArcLengths()
	return out
}

func (fp Stroke) boundingBox() Rect {
	re := EmptyRect()
	for _, cu := range fp.Curves {
		re.Union(cu.boundingBox())
	}
	return re
}

func (fp Stroke) controlBox() Rect {
	re := fp.Curves[0].controlBox()
	for _, cu := range fp.Curves {
		re.Union(cu.controlBox())
	}
	return re
}

func (fp Stroke) String() string {
	curves := make([]string, len(fp.Curves))
	for i, c := range fp.Curves {
		curves[i] = c.String()
	}
	return fmt.Sprintf("{Curves: []Bezier{%s}, ArcLengths: %#v}", strings.Join(curves, ", "), fp.ArcLengths)
}

func mapFromTo(U, V Rect) Trans {
	// rescale U to V
	startU, endU := U.UL, U.LR
	startV, endV := V.UL, V.LR

	Lu, Lv := distP(startU, endU), distP(startV, endV)
	s := Lv / Lu
	// align barycenter of scaled U
	Bu := startU.Add(endU).ScaleTo(0.5 * s)
	Bv := startV.Add(endV).ScaleTo(0.5)

	t := Bv.Sub(Bu)

	return Trans{s, t}
}

// scale apply [tr] to all the bezier curves, returning a new shape
func (fp Stroke) scale(tr Trans) Stroke {
	out := Stroke{
		Curves: make([]Bezier, len(fp.Curves)),
		// note that tr preserve lengths
		ArcLengths: fp.ArcLengths,
	}
	for i, c := range fp.Curves {
		out.Curves[i] = c.Scale(tr)
	}
	return out
}

func startEnd(cu []Bezier) Rect {
	return Rect{cu[0].P0, cu[len(cu)-1].P3}
}

// rescale U to V
func distanceFootprints(U, V Stroke) Fl {
	tr := mapFromTo(startEnd(U.Curves), startEnd(V.Curves))
	U = U.scale(tr)

	return distanceFootprintNoScale(U, V)
}

// distanceFootprintNoScale returns the distance between U and V,
// without rescaling step
func distanceFootprintNoScale(U, V Stroke) Fl {
	c1s, c2s, ok := adjustFootprints(U, V)
	if !ok {
		return Inf
	}

	var (
		totalDist   Fl
		totalLength Fl
	)
	for i := range c1s {
		c1 := c1s[i]
		c2 := c2s[i]

		d := c1.distance(c2)
		length := c1.arcLength()

		fmt.Println("Curve", i, d, length)

		totalDist += d * length
		totalLength += length
	}

	return totalDist / totalLength
}

// adjustFootprints assume U has been scaled to match V
func adjustFootprints(U, V Stroke) (c1s, c2s []Bezier, ok bool) {
	// rescale both U and V to a reference size so that error values are
	// comparable across the database
	cbox := U.controlBox()
	h, w := cbox.Height(), cbox.Width()
	// handle linear sections
	s := Max(Max(h, w), 1)

	tr := Trans{Scale: 20 / s}
	U = U.scale(tr)
	V = V.scale(tr)

	// compute the common subdivision
	split1, split2, ok := mapBetweenArcLengths(U.ArcLengths, V.ArcLengths)
	if !ok {
		return nil, nil, false
	}
	c1s, c2s = U.split(split1), V.split(split2)

	return c1s, c2s, true
}

// map curve index to offsets t in the curve ( between 0 and 1)
type splitMap map[int][]Fl

// establish a mapping from the two subdivisions :
//   - close enough values are mapped
//   - when necessary, range are split
//
// after applying the two list of split returned,
// the footprints will have the same number of curves
func mapBetweenArcLengths(a1, a2 []Fl) (s1, s2 splitMap, ok bool) {
	i1, i2 := 0, 0
	s1, s2 = make(splitMap), make(splitMap)
	for i1 < len(a1) && i2 < len(a2) {
		const gapWidth = 0.05
		v1, v2 := a1[i1], a2[i2]
		// are the values roughly the same ?
		if abs(v1-v2) < gapWidth {
			// nothing to do
			i1++
			i2++
			continue
		} else {
			var prev1, prev2 Fl
			if i1 != 0 {
				prev1 = a1[i1-1]
			}
			if i2 != 0 {
				prev2 = a2[i2-1]
			}

			// check if there is a significant gap,
			// to be split up
			if v1 < v2 {
				// try to split [prev2, v1, v2]
				if abs(v1-prev2) > gapWidth && abs(v2-v1) > gapWidth {
					// split and map the value from [prev2, v2] to [0,1]
					t1 := (v1 - prev2) / (v2 - prev2)
					s2[i2] = append(s2[i2], t1)
					i1++
				} else { // consider the shapes are incompatible
					return nil, nil, false
				}
			} else { // same
				// try to split [prev1, v2, v1]
				if abs(v2-prev1) > gapWidth && abs(v1-v2) > gapWidth {
					// split and map the value from [prev1, v1] to [0,1]
					t2 := (v2 - prev1) / (v1 - prev1)
					s1[i1] = append(s1[i1], t2)
					i2++
				} else { // consider the shapes are incompatible
					return nil, nil, false
				}
			}
		}
	}

	return s1, s2, true
}

func (fp Stroke) split(splits splitMap) []Bezier {
	if len(splits) == 0 {
		return fp.Curves
	}
	var out []Bezier
	for i, c := range fp.Curves {
		sp := splits[i]
		if len(sp) == 0 { // no split
			out = append(out, c)
		} else {
			sp = append(sp, 1)
			for j, t := range sp {
				tstart, tend := Fl(0), t
				if j != 0 {
					tstart = sp[j-1]
				}
				out = append(out, c.splitBetween(tstart, tend))
			}
		}
	}
	return out
}

// ----------------- extension to whole symbols -----------------

// Footprint stores a simplfied representation
// of a [Symbol].
type Footprint struct {
	Strokes []Stroke
}

func newSymbolFootprint(sy Symbol) Footprint {
	strokes := make([]Stroke, len(sy))
	for i, shape := range sy {
		strokes[i] = newFp(shape)
	}
	out := Footprint{Strokes: strokes}

	return out
}

// controlBox returns the union of the control box of each shape.
// It is a cheap approximation of the bounding box.
func (sf Footprint) controlBox() Rect {
	out := EmptyRect()
	for _, sh := range sf.Strokes {
		out.Union(sh.controlBox())
	}
	return out
}

// BoundingBox returns the union of the bounding box of each shape.
func (sf Footprint) BoundingBox() Rect {
	out := EmptyRect()
	for _, sh := range sf.Strokes {
		out.Union(sh.boundingBox())
	}
	return out
}

// assuming len(store) > len(input), only compares the first common strokes
func distanceSymbolsCompatible(store, input Footprint) Fl {
	nbStrokes := len(input.Strokes)

	// use the same number of components as the input
	if len(store.Strokes) < nbStrokes { // ignore this entry
		return Inf
	}
	store.Strokes = store.Strokes[:nbStrokes] // only mutate the local [store]

	return distanceSymbolsExact(store, input)
}

func distanceStrokes(s1, s2 []Stroke) Fl {
	var totalDistance Fl
	for i := range s1 {
		fpU, fpV := s1[i], s2[i]

		// accept a curve drawn in opposite order
		if debugMode {
			fmt.Println("Distance between stroke (regular)")
		}
		d1 := distanceFootprintNoScale(fpU, fpV)

		if debugMode {
			fmt.Println("Distance between stroke (reversed)")
		}
		d2 := distanceFootprintNoScale(fpU.reverse(), fpV)

		d := Min(d1, d2)

		totalDistance += d
	}

	return totalDistance / Fl(len(s1))
}

// distanceSymbolsExact compare two footprints for whole symbols
// is always return infinity if the symbols have not the same length
func distanceSymbolsExact(U, V Footprint) Fl {
	if len(U.Strokes) != len(V.Strokes) {
		return Inf
	}

	aspectRatioU := U.BoundingBox().Width() / U.BoundingBox().Height()
	aspectRatioV := V.BoundingBox().Width() / V.BoundingBox().Height()
	fmt.Println(aspectRatioU, aspectRatioV)

	// rescale U to V, with the same transformation for
	// each shapes
	tr := mapFromTo(U.controlBox(), V.controlBox())

	// build a copy with scaling applied
	Uscaled := make([]Stroke, len(U.Strokes))
	for i, s := range U.Strokes {
		Uscaled[i] = s.scale(tr) // apply the scale
	}

	dist := distanceStrokes(Uscaled, V.Strokes)

	// some symbols may be written in any order
	// to avoid complicating too much, we only try permutation
	// for two-strokes symbols, with one curve
	if len(Uscaled) == 2 && len(Uscaled[0].Curves) == 1 && len(Uscaled[1].Curves) == 1 &&
		len(V.Strokes[0].Curves) == 1 && len(V.Strokes[1].Curves) == 1 {
		permutated := []Stroke{Uscaled[1], Uscaled[0]}
		if d := distanceStrokes(permutated, V.Strokes); d < dist {
			dist = d
		}
	}

	return dist
}
