package layout

import (
	"fmt"
	"strings"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

const debugMode = true

type Fl = sy.Fl

// Record stores the raw user input
type Record []sy.Shape

func (re Record) String() string {
	var builder strings.Builder
	for _, s := range re {
		builder.WriteString(s.String())
		builder.WriteString(",\n")
	}

	return fmt.Sprintf(`Record{
		%s
	}`, builder.String())
}

// split returns the symbol drawn before the last connex shape and
// the last connex compoment drawn
func (rec Record) split() (sy.Symbol, sy.Shape) {
	if len(rec) == 0 {
		return nil, nil
	}
	return sy.Symbol(rec[0 : len(rec)-1]), rec[len(rec)-1]
}

// return the distance between the closest point between [u] and [v]
// This NOT a measure of similarity
func closestPointDistance(u, v sy.Shape) Fl {
	best := sy.Inf
	for _, pu := range u {
		for _, pv := range v {
			if d := pu.Sub(pv).NormSquared(); d < best {
				best = d
			}
		}
	}
	return sy.Sqrt(best)
}

func (rec Record) Identify(store *sy.Store) rune {
	if toMatch, ok := rec.isSeparated(); ok { // easy case : only use the last stroke

		if debugMode {
			fmt.Println("Identify record : isSeparated -> last stroke matched")
		}

		r, _ := store.Lookup(toMatch.Footprint(), sy.Rect{})
		return r
	}
	// here, len(rec) > 1
	wholeFootprint := sy.Symbol(rec).Footprint()
	previous, last := wholeFootprint.Strokes[:len(wholeFootprint.Strokes)-1], wholeFootprint.Strokes[len(wholeFootprint.Strokes)-1]
	previousFooprint, lastFootprint := sy.Footprint{Strokes: previous}, sy.Footprint{Strokes: []sy.Stroke{last}}

	if isMerged(previous, last) {

		if debugMode {
			fmt.Println("Identify record : isMerged -> whole symbol matched")
		}

		r, _ := store.Lookup(wholeFootprint, sy.Rect{})
		return r
	}

	// special for points
	if len(last.Curves) == 1 {
		if point, ok := last.Curves[0].IsPoint(); ok {
			// decide to match the whole symbol base on the X value
			bbox := previousFooprint.BoundingBox()
			fmt.Println("point, ", bbox.LR.X, point.X)
			if bbox.LR.X+2 >= point.X {

				if debugMode {
					fmt.Println("Identify record : point -> whole symbol matched")
				}

				r, _ := store.Lookup(wholeFootprint, sy.Rect{})
				return r
			}
		}
	}

	// here we are not sure : it could be two distinct symbols
	// or only one sligtly separated like x, ℝ, ...
	//
	// to disambiguate, we perform to lookups and compare errors

	// lookup with the whole symbol
	rWhole, errWhole := store.Lookup(wholeFootprint, sy.Rect{})

	// lookup with separated symbol
	_, errPrevious := store.Lookup(previousFooprint, sy.Rect{})
	rLast, errLast := store.Lookup(lastFootprint, sy.Rect{})

	fmt.Println(errPrevious, errLast, errWhole)

	// it always easier to match separate parts : compense a bit
	if sy.Max(errPrevious, errLast) < 0.9*errWhole {

		if debugMode {
			fmt.Println("Identify record : after lookup -> last stroke matched")
		}

		return rLast
	}

	if debugMode {
		fmt.Println("Identify record : after lookup -> whole stroke matched")
	}

	return rWhole
}

// return true if the last stroke has one intersection
// with the others
func isMerged(previous []sy.Stroke, last sy.Stroke) bool {
	if len(last.Curves) != 1 {
		return false
	}
	seg := last.Curves[0]
	if !seg.IsRoughlyLinear() {
		return false
	}
	start, end := seg.P0, seg.P3
	for _, stroke := range previous {
		for _, cu := range stroke.Curves {
			if cu.IntersectsSegment(start, end) {
				return true
			}
		}
	}
	return false
}

// isSeparated returns true if we are certain only the last
// stroke should be used when matching runes
// is always returns true if len(rec) == 1
func (rec Record) isSeparated() (sy.Symbol, bool) {
	const splitWidth = 5

	if len(rec) == 1 { // only one stroke
		return sy.Symbol(rec), true
	}

	previous, last := rec.split()

	// we consider that compound symbol always have strokes with overlapping X values
	previousBbox, lastBbox := previous.Union().BoundingBox(), last.BoundingBox()
	if previousBbox.LR.X+splitWidth < lastBbox.UL.X || // previous then last
		lastBbox.LR.X+splitWidth < previousBbox.UL.X { // last then previous
		return sy.Symbol{last}, true
	}

	return nil, false
}

// Recorder is used to record the shapes
// drawn by the user
type Recorder struct {
	// Record is the accumulated Record
	Record Record

	currentShape sy.Shape
	inShape      bool
}

func (rec *Recorder) DropButLast() {
	if len(rec.Record) == 0 {
		return
	}
	rec.Record = Record{rec.Record[len(rec.Record)-1]}
}

// Reset clears the current state of the `Recorder`
func (rec *Recorder) Reset() {
	rec.inShape = false
	rec.Record = nil
	rec.currentShape = nil
}

// StartShape starts the recording of a new connex shape.
func (rec *Recorder) StartShape() {
	rec.inShape = true
	rec.currentShape = sy.Shape{}
}

// EndShape ends the recording of the current connex shape.
// Empty shapes are discarded.
func (rec *Recorder) EndShape() {
	rec.inShape = false
	if current := rec.currentShape; len(current) != 0 {
		rec.Record = append(rec.Record, current)
	}
	rec.currentShape = nil
}

// AddToShape adds a point to the shape, if one has started
func (rec *Recorder) AddToShape(pos sy.Pos) {
	if rec.inShape {
		rec.currentShape = append(rec.currentShape, pos)
	}
}
