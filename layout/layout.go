// Package layout implements a data structure
// representing a math expression and its layout,
// used to recognize and correctly insert a user input
package layout

import (
	"fmt"
	"strings"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

const (
	EMWidth         float32 = 30.
	EMHeight        float32 = 60.
	EMBaselineRatio float32 = 0.66 // from the top
)

// grapheme is one symbol input, with its resolved Unicode value
type grapheme struct {
	Symbol sy.Footprint
	Char   rune // the resolved rune for the symbol
}

func (nc grapheme) Grapheme() grapheme { return nc }

// block is a component of an [Expression] coming
// from a symbol (such as a letter, a digit, a math operator),
// and including potential children.
type block interface {
	// Grapheme returns the main grapheme for the block,
	// which usually has the visual appearence of the block with empty
	// children.
	Grapheme() grapheme

	// Children returns a list of the children nodes, which are
	// always valid pointer, but possibly empty nodes.
	Children() []*Node

	// laTeX returns the LaTeX code for this block and its children
	laTeX() string
}

func blockBox(bl block, includeMargin bool) sy.Rect {
	own := bl.Grapheme().Symbol.BoundingBox()
	for _, child := range bl.Children() {
		childContext := child.context(includeMargin)
		own.Union(childContext.Box)
	}
	return own
}

// Node represents one level of an expression.
// An empty node is used when no symbols have been written yet.
type Node struct {
	minBox   sy.Rect
	baseline sy.Fl

	blocks []block
}

func newNode(box sy.Rect, baseline Fl) *Node {
	return &Node{minBox: box, baseline: baseline}
}

// include the potential children
// if includeMargin is true, the minimal box is used, and margin after the children is added
func (n *Node) context(includeMargin bool) sy.Context {
	// compute the union of the children
	childrenBbox := sy.EmptyRect()
	for _, block := range n.blocks {
		box := blockBox(block, includeMargin)
		childrenBbox.Union(box)
	}

	if includeMargin {
		// and add margin for new inputs
		childrenBbox.LR.X += EMWidth
		// make sure it includes at least minBox
		childrenBbox.Union(n.minBox)
	}

	return sy.Context{Box: childrenBbox, Baseline: n.baseline}
}

// return the concatenation of the latex for each block
func (n *Node) latex() string {
	chunks := make([]string, len(n.blocks))
	for i, b := range n.blocks {
		chunks[i] = b.laTeX()
	}
	return strings.Join(chunks, "")
}

// insertAt insert [bl] at Blocks[index]
func (n *Node) insertAt(bl block, index int) {
	n.blocks = append(n.blocks, nil /* use the zero value of the element type */)
	copy(n.blocks[index+1:], n.blocks[index:])
	n.blocks[index] = bl
}

// Line represents one line of text, wrote
// in the top level scope.
type Line struct {
	root Node
	// To handle Symbols made of several Shapes,
	// cursor stores the current Grapheme
	// cursor *grapheme
}

func NewLine(rect sy.Rect) *Line {
	return &Line{root: *newNode(rect, baselineFromRect(rect))}
}

func (li *Line) Symbols() (out []sy.Footprint) {
	var aux func(node *Node)
	aux = func(node *Node) {
		for _, char := range node.blocks {
			out = append(out, char.Grapheme().Symbol)
			for _, child := range char.Children() {
				aux(child)
			}
		}
	}

	aux(&li.root)

	return out
}

// Contexts returns all the possible input areas.
// This is to be used for debugging purposes.
func (li *Line) Contexts() (out []sy.Rect) {
	var aux func(node *Node)
	aux = func(node *Node) {
		out = append(out, node.context(true).Box)
		for _, char := range node.blocks {
			out = append(out, blockBox(char, true))
			// recurse
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
	grapheme

	indice, exponent *Node
}

func newRegularChar(gr grapheme) *regularChar {
	bb := gr.Symbol.BoundingBox()
	xLeft := bb.LR.X - 0.1*bb.Width()

	// the height is fixed to a proportion of the EM square
	height := EMHeight * 0.4

	// adjust the baseline of the exponent scope to the height of the char
	baselineExponent := bb.UL.Y
	exponent := rectForBaseline(xLeft, EMWidth, height, baselineExponent)

	// adjust the baseline of the indice scope just under the char
	baselineIndice := bb.LR.Y + 0.1*EMHeight
	indice := rectForBaseline(xLeft, EMWidth, height, baselineIndice)

	return &regularChar{grapheme: gr, exponent: newNode(exponent, baselineExponent), indice: newNode(indice, baselineIndice)}
}

func (r *regularChar) Children() []*Node { return []*Node{r.indice, r.exponent} }

// return a rect such that its baseline (top + height*EMBaselineRatio) is at baseline
func rectForBaseline(x, width, height, baseline float32) sy.Rect {
	top := baseline - height*EMBaselineRatio
	return sy.Rect{
		UL: sy.Pos{X: x, Y: top},
		LR: sy.Pos{X: x + width, Y: top + height},
	}
}

func baselineFromRect(r sy.Rect) Fl {
	return r.UL.Y + r.Height()*EMBaselineRatio
}

func (r regularChar) laTeX() string {
	out := string(r.grapheme.Char)
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
	grapheme

	num, den *Node
}

func newfracOperator(gr grapheme) *fracOperator {
	// the width is given by the fraction itself

	bbox := gr.Symbol.BoundingBox()
	fracTop, fracBottom := bbox.UL.Y, bbox.LR.Y

	// enlarge the num height
	numHeight := EMHeight * 0.9
	numBox := sy.Rect{UL: sy.Pos{X: bbox.UL.X, Y: fracTop - numHeight}, LR: sy.Pos{X: bbox.LR.X, Y: fracTop}}
	numBaseline := baselineFromRect(numBox)

	// enlarge the den height
	denHeight := EMHeight * 0.9
	denBox := sy.Rect{UL: sy.Pos{X: bbox.UL.X, Y: fracBottom}, LR: sy.Pos{X: bbox.LR.X, Y: fracBottom + denHeight}}
	denBaseline := baselineFromRect(denBox)

	return &fracOperator{grapheme: gr, num: newNode(numBox, numBaseline), den: newNode(denBox, denBaseline)}
}

func (f *fracOperator) Children() []*Node { return []*Node{f.num, f.den} }

func (r fracOperator) laTeX() string {
	num, den := r.num.latex(), r.den.latex()
	return fmt.Sprintf(`\frac{%s}{%s}`, num, den)
}

type sumOperator struct {
	grapheme

	from, to *Node
	content  *Node
}

func (s *sumOperator) Children() []*Node { return []*Node{s.from, s.to, s.content} }

// TODO:
func (r sumOperator) laTeX() string { return "" }

type prodOperator sumOperator

func (p *prodOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

// TODO:
func (r prodOperator) laTeX() string { return "" }

type integralOperator sumOperator

func (p *integralOperator) Children() (out []*Node) { return (*sumOperator)(p).Children() }

// TODO:
func (r integralOperator) laTeX() string { return "" }
