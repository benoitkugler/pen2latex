package symbols

import (
	"fmt"
	"strings"
)

// [Lookup] performs approximate matching by finding
// the closest symbol to [input] in the database and returning its rune.
//
// It will return 0 if the store is empty or if no symbol matches [input].
//
// # Matching overview
//
// To match an input [Record] against the database,
// we perform the following steps :
//   - for each [Shape] in the record, segment it into Bezier curves, yielding a [ShapeFootprint]
//   - for each symbol entry in the database, compute the distance between its footprint and the input
//   - TODO: disambiguate results using the size of the surrounding context
func (db *Store) Lookup(input Symbol, context Rect) rune {
	var (
		bestIndex    int = -1
		bestDistance     = Inf
	)

	inputFootprint := input.Footprint()
	nbShapes := len(input)

	for i, entry := range db.Symbols {
		// use the same number of components as the input
		if len(entry.Footprint) < nbShapes { // ignore this entry
			continue
		}
		distance := distanceSymbols(entry.Footprint[:nbShapes], inputFootprint)
		if distance < bestDistance {
			bestDistance = distance
			bestIndex = i
		}
	}

	if bestIndex == -1 {
		return 0
	}
	return db.Symbols[bestIndex].R
}

// Footprint builds the footprint of the symbol
func (sy Symbol) Footprint() Footprint { return newSymbolFootprint(sy) }

// ShapeFP stores a simplified representation of one
// [Shape]
type ShapeFP struct {
	Curves     []Bezier `json:"c"`
	ArcLengths []Fl     `json:"a"` // between 0 and 1, starts after the first part and ends at 1
}

func newFp(points Shape) ShapeFP {
	if len(points) < 4 {
		return ShapeFP{}
	}

	// fit and regularize
	curves := mergeSimilarCurves(fitCubicBeziers(points))

	// compute arc lengths
	arcLengths := make([]Fl, len(curves))
	var totalLength Fl
	for i, part := range curves {
		L := part.arcLength()
		totalLength += L
		arcLengths[i] = totalLength
	}

	// normalize
	for i := range arcLengths {
		arcLengths[i] /= totalLength
	}

	return ShapeFP{Curves: curves, ArcLengths: arcLengths}
}

func (fp ShapeFP) boundingBox() Rect {
	re := EmptyRect()
	for _, cu := range fp.Curves {
		re.Union(cu.boundingBox())
	}
	return re
}

func (fp ShapeFP) controlBox() Rect {
	re := fp.Curves[0].controlBox()
	for _, cu := range fp.Curves {
		re.Union(cu.controlBox())
	}
	return re
}

func (fp ShapeFP) String() string {
	curves := make([]string, len(fp.Curves))
	for i, c := range fp.Curves {
		curves[i] = c.String()
	}
	return fmt.Sprintf("{curves: []Bezier{%s}, arcLengths: %#v}", strings.Join(curves, ", "), fp.ArcLengths)
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
func (fp ShapeFP) scale(tr Trans) ShapeFP {
	out := ShapeFP{
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
func distanceFootprints(U, V ShapeFP) Fl {
	tr := mapFromTo(startEnd(U.Curves), startEnd(V.Curves))
	U = U.scale(tr)

	return distanceFootprintNoScale(U, V)
}

// distanceFootprintNoScale returns the distance between U and V,
// without rescaling step
func distanceFootprintNoScale(U, V ShapeFP) Fl {
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

		totalDist += d * length
		totalLength += length
	}

	return totalDist / totalLength
}

// adjustFootprints assume U has been scaled to match V
func adjustFootprints(U, V ShapeFP) (c1s, c2s []Bezier, ok bool) {
	// rescale both U and V to a reference size so that error values are
	// comparable across the database
	cbox := U.controlBox()
	h, w := cbox.Height(), cbox.Width()
	if w < 0.1 { // handle linear sections
		w = 1
	}
	if h < 0.1 { // handle linear sections
		h = 1
	}
	tr := Trans{Scale: 20 / Sqrt(h*w)}
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
		v1, v2 := a1[i1], a2[i2]
		// are the values roughly the same ?
		if abs(v1-v2) < 0.1 {
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
			const gapWidth = 0.1
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

func (fp ShapeFP) split(splits splitMap) []Bezier {
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
type Footprint []ShapeFP

func newSymbolFootprint(sy Symbol) Footprint {
	out := make(Footprint, len(sy))
	for i, shape := range sy {
		out[i] = newFp(shape)
	}
	return out
}

// controlBox returns the union of the control box of each shape.
// It is a cheap approximation of the bounding box.
func (sf Footprint) controlBox() Rect {
	out := EmptyRect()
	for _, sh := range sf {
		out.Union(sh.controlBox())
	}
	return out
}

// BoundingBox returns the union of the bounding box of each shape.
func (sf Footprint) BoundingBox() Rect {
	out := EmptyRect()
	for _, sh := range sf {
		out.Union(sh.boundingBox())
	}
	return out
}

// distanceSymbols compare two footprints for whole symbols
// is always return infinity if the symbols have not the same length
func distanceSymbols(U, V Footprint) Fl {
	if len(U) != len(V) {
		return Inf
	}

	// rescale U to V, with the same transformation for
	// each shapes
	tr := mapFromTo(U.controlBox(), V.controlBox())

	var totalDistance Fl
	for i := range U {
		fpU, fpV := U[i], V[i]
		fpU = fpU.scale(tr) // apply the scale
		d := distanceFootprintNoScale(fpU, fpV)
		totalDistance += d
	}

	return totalDistance
}
