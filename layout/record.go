package layout

import (
	"fmt"
	"strings"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

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

// InferSymbol decides which part of the record should be used
// to match a rune.
func (rec Record) InferSymbol() sy.Symbol {
	const penWidth = 4

	last, shape := rec.split()
	// we consider that compound symbol always have adjacent strokes
	if d := closestPointDistance(last.Union(), shape); d < penWidth { // we have a compound
		return sy.Symbol(rec)
	}

	return sy.Symbol{shape}
}

// Recorder is used to record the shapes
// drawn by the user
type Recorder struct {
	// Record is the accumulated Record
	Record Record

	currentShape sy.Shape
	inShape      bool
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
