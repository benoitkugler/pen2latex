// Package layout implements a data structure
// representing a math expression and its layout,
// used to recognize and correctly insert a user input
package layout

import (
	"github.com/benoitkugler/pen2latex/symbols"
)

const (
	EMWidth         float32 = 20.
	EMHeight        float32 = 50.
	EMBaselineRatio float32 = 0.66 // from the top
)

// Scope is an area in the editor where potential new
// symbols may be drawn. Its area is never empty :
// it contains already written input or a blank area
// if the input is empty.
type Scope = symbols.Rect

// Grapheme is one symbol input, with its resolved Unicode value
type Grapheme struct {
	Char   rune // the resolved rune for the symbol
	Symbol symbols.Symbol
}

func (nc *Grapheme) Content() *Grapheme { return nc }

// Block is a component of an [Expression] coming
// from a symbol (such as a letter, a digit, a math operator),
// and including potential children.
type Block interface {
	// Content returns the main grapheme for the block,
	// which usually has the visual appearence of the block with empty
	// children.
	Content() *Grapheme

	// Children returns a list of the children nodes, which are
	// always valid pointer, but possibly empty nodes.
	Children() []*Node

	// Scopes returns a list of the children scopes, in the same order
	// as [Children]
	Scopes() []Scope
}

// Node represents one level of an expression.
// An empty node is used when no symbols have been written yet.
type Node struct {
	Blocks []Block
}

// returns EmptyRect if n is empty
func (n *Node) extendedBox() symbols.Rect {
	bbox := symbols.EmptyRect()
	for _, char := range n.Blocks {
		bbox.Union(extendedBox(char))
	}
	return bbox
}

// extendedBox return the extended bounding box,
// including the glyph and its children
func extendedBox(ch Block) symbols.Rect {
	own := ch.Content().Symbol.Union().BoundingBox()
	for _, child := range ch.Scopes() {
		own.Union(child)
	}
	return own
}

// Line represents one line of text, wrote
// in the top level scope.
type Line struct {
	root Node
	// To handle Symbols made of several Shapes,
	// cursor stores the current Grapheme
	cursor *Grapheme
}

func (li *Line) Symbols() (out []symbols.Symbol) {
	var aux func(node *Node)
	aux = func(node *Node) {
		for _, char := range node.Blocks {
			out = append(out, char.Content().Symbol)
			for _, child := range char.Children() {
				aux(child)
			}
		}
	}

	aux(&li.root)

	return out
}

func (li *Line) Scopes() (out []Scope) {
	var aux func(node *Node)
	aux = func(node *Node) {
		for _, char := range node.Blocks {
			out = append(out, char.Scopes()...)
			for _, child := range char.Children() {
				aux(child)
			}
		}
	}

	aux(&li.root)

	return out
}

var (
	_ Block = (*regularChar)(nil)
	_ Block = (*fracOperator)(nil)
	_ Block = (*sumOperator)(nil)
	_ Block = (*prodOperator)(nil)
	_ Block = (*integralOperator)(nil)
)

// regularChar is a simple character which
// has the two default scopes: indice and exponent
type regularChar struct {
	Grapheme

	indice, exponent *Node
}

func (r *regularChar) Children() []*Node { return []*Node{r.indice, r.exponent} }

// return a rect such that its baseline (top + height*EMBaselineRatio) is at baseline
func rectForBaseline(x, width, height, baseline float32) symbols.Rect {
	top := baseline - height*EMBaselineRatio
	return symbols.Rect{
		UL: symbols.Pos{X: x, Y: top},
		LR: symbols.Pos{X: x + width, Y: top + height},
	}
}

// Scopes return two scopes : the indice area and
// the exponent area
func (r *regularChar) Scopes() []Scope {
	u := r.Grapheme.Symbol.Union()
	bb := u.BoundingBox()
	xLeft := bb.LR.X - 0.1*bb.Width()

	// the height is fixed to a proportion of the EM square
	height := EMHeight * 0.5

	// enlarge by the current exponent bbox width
	exponentWith := EMWidth
	if expBB := r.exponent.extendedBox(); !expBB.IsEmpty() {
		exponentWith += bb.Width()
	}
	// adjust the baseline of the exponent scope to the height of the char
	exponent := rectForBaseline(xLeft, exponentWith, height, bb.UL.Y)

	// enlarge by the current exponent bbox width
	indiceWith := EMWidth
	if expBB := r.indice.extendedBox(); !expBB.IsEmpty() {
		indiceWith += bb.Width()
	}
	// adjust the baseline of the indice scope just under the char
	indice := rectForBaseline(xLeft, indiceWith, height, bb.LR.Y+0.1*EMHeight)

	return []Scope{indice, exponent}
}

type fracOperator struct {
	Grapheme

	num, den *Node
}

func (f *fracOperator) Children() []*Node { return []*Node{f.num, f.den} }

func (fracOperator) Scopes() []Scope {
	// TODO:
	return nil
}

type sumOperator struct {
	Grapheme

	from, to *Node
	content  *Node
}

func (s *sumOperator) Children() []*Node { return []*Node{s.from, s.to, s.content} }

func (sumOperator) Scopes() []Scope {
	// TODO:
	return nil
}

type prodOperator sumOperator

func (p *prodOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

func (prodOperator) Scopes() []Scope {
	// TODO:
	return nil
}

type integralOperator sumOperator

func (p *integralOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

func (integralOperator) Scopes() []Scope {
	// TODO:
	return nil
}
