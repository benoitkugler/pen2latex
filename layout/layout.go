// Package layout implements a data structure
// representing a math expression and its layout,
// used to recognize and correctly insert a user input
package layout

import (
	"fmt"
	"strings"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

// Scope is an area in the editor where potential new
// symbols may be drawn. Its area is never empty :
// it contains already written input or a blank area
// if the input is empty.
type Scope = sy.Rect

// Grapheme is one symbol input, with its resolved Unicode value
type Grapheme struct {
	Char   rune // the resolved rune for the symbol
	Symbol sy.Symbol
}

func (nc *Grapheme) Content() *Grapheme { return nc }

// block is a component of an [Expression] coming
// from a symbol (such as a letter, a digit, a math operator),
// and including potential children.
type block interface {
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

	// laTeX returns the LaTeX code for this block and its children
	laTeX() string
}

// Node represents one level of an expression.
// An empty node is used when no symbols have been written yet.
type Node struct {
	Blocks []block
}

func (n *Node) latex() string {
	chunks := make([]string, len(n.Blocks))
	for i, b := range n.Blocks {
		chunks[i] = b.laTeX()
	}
	return strings.Join(chunks, "")
}

// insertAt insert [bl] at Blocks[index]
func (n *Node) insertAt(bl block, index int) {
	n.Blocks = append(n.Blocks, nil /* use the zero value of the element type */)
	copy(n.Blocks[index+1:], n.Blocks[index:])
	n.Blocks[index] = bl
}

// returns EmptyRect if n is empty
func (n *Node) extendedBox() sy.Rect {
	bbox := sy.EmptyRect()
	for _, char := range n.Blocks {
		bbox.Union(extendedBox(char))
	}
	return bbox
}

// extendedBox return the extended bounding box,
// including the glyph and its children
func extendedBox(ch block) sy.Rect {
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

func (li *Line) Symbols() (out []sy.Symbol) {
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

// LaTeX returns the LaTeX code deduced from the current drawings
func (li *Line) LaTeX() string { return li.root.latex() }

var (
	_ block = (*regularChar)(nil)
	_ block = (*fracOperator)(nil)
	_ block = (*sumOperator)(nil)
	_ block = (*prodOperator)(nil)
	_ block = (*integralOperator)(nil)
)

// regularChar is a simple character which
// has the two default scopes: indice and exponent
type regularChar struct {
	Grapheme

	indice, exponent *Node
}

func (r *regularChar) Children() []*Node { return []*Node{r.indice, r.exponent} }

// return a rect such that its baseline (top + height*EMBaselineRatio) is at baseline
func rectForBaseline(x, width, height, baseline float32) sy.Rect {
	top := baseline - height*sy.EMBaselineRatio
	return sy.Rect{
		UL: sy.Pos{X: x, Y: top},
		LR: sy.Pos{X: x + width, Y: top + height},
	}
}

// Scopes return two scopes : the indice area and
// the exponent area
func (r *regularChar) Scopes() []Scope {
	u := r.Grapheme.Symbol.Union()
	bb := u.BoundingBox()
	xLeft := bb.LR.X - 0.1*bb.Width()

	// the height is fixed to a proportion of the EM square
	height := sy.EMHeight * 0.4

	// enlarge by the current exponent bbox width
	exponentWith := sy.EMWidth
	if expBB := r.exponent.extendedBox(); !expBB.IsEmpty() {
		exponentWith += bb.Width()
	}
	// adjust the baseline of the exponent scope to the height of the char
	exponent := rectForBaseline(xLeft, exponentWith, height, bb.UL.Y)

	// enlarge by the current exponent bbox width
	indiceWith := sy.EMWidth
	if expBB := r.indice.extendedBox(); !expBB.IsEmpty() {
		indiceWith += bb.Width()
	}
	// adjust the baseline of the indice scope just under the char
	indice := rectForBaseline(xLeft, indiceWith, height, bb.LR.Y+0.1*sy.EMHeight)

	return []Scope{indice, exponent}
}

func (r regularChar) laTeX() string {
	out := string(r.Grapheme.Char)
	indice := r.indice.latex()
	exponent := r.exponent.latex()
	if indice != "" {
		out += fmt.Sprintf("_{%s}", indice)
	}
	if exponent != "" {
		out += fmt.Sprintf("^{%s}", exponent)
	}
	return out
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

// TODO:
func (r fracOperator) laTeX() string { return "" }

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

// TODO:
func (r sumOperator) laTeX() string { return "" }

type prodOperator sumOperator

func (p *prodOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

func (prodOperator) Scopes() []Scope {
	// TODO:
	return nil
}

// TODO:
func (r prodOperator) laTeX() string { return "" }

type integralOperator sumOperator

func (p *integralOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

func (integralOperator) Scopes() []Scope {
	// TODO:
	return nil
}

// TODO:
func (r integralOperator) laTeX() string { return "" }
