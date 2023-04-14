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
	EMBaselineRatio float32 = 0.7 // from the top
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

func blockBoxes(bl block) (inner, outer sy.Rect) {
	inner = bl.Grapheme().Symbol.BoundingBox()
	outer = inner // copy
	for _, child := range bl.Children() {
		ic, oc := child.boxes()
		inner.Union(ic)
		outer.Union(oc)
	}
	return
}

// Node represents one level of an expression.
// An empty node is used when no symbols have been written yet.
type Node struct {
	height   sy.HeightGrid // fixed height
	initialX Fl            // used when the node is empty

	blocks []block
}

func newNode(height sy.HeightGrid, x Fl) *Node {
	return &Node{height: height, initialX: x}
}

// boxes return two boxes delimiting the content of the node
// and its children
// [inner] only accounts for the actual drawn symbols,
// whereas [outer] adds a placeholder space
func (n *Node) boxes() (inner, outer sy.Rect) {
	// compute the union of the children

	if len(n.blocks) == 0 {
		// use the initial X
		inner = sy.Rect{
			UL: sy.Pos{X: n.initialX, Y: n.height.Ymin},
			LR: sy.Pos{X: n.initialX, Y: n.height.Ymax},
		}
	} else {
		inner = sy.EmptyRect()
	}
	outer = inner // copy
	for _, block := range n.blocks {
		innerBox, outerBox := blockBoxes(block)
		inner.Union(innerBox)
		outer.Union(outerBox)
	}

	// ensure [outer] box has a margin with respect to inner
	if withMargin := inner.LR.X + EMWidth; outer.LR.X < withMargin {
		outer.LR.X = withMargin
	}
	return
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
	// cursor stores the last block inserted
	cursor *block
}

func NewLine(rect sy.Rect) *Line {
	return &Line{root: *newNode(baselineFromHeight(rect.UL.Y, rect.LR.Y), rect.UL.X)}
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

// // Contexts returns all the possible input areas.
// // This is to be used for debugging purposes.
// func (li *Line) Contexts() (out []sy.Rect) {
// 	var aux func(node *Node)
// 	aux = func(node *Node) {
// 		out = append(out, node.context(true).Box)
// 		for _, char := range node.blocks {
// 			out = append(out, blockBox(char, true))
// 			// recurse
// 			for _, child := range char.Children() {
// 				aux(child)
// 			}
// 		}
// 	}

// 	aux(&li.root)

// 	return out
// }

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
	xLeft := bb.LR.X - 0.2*bb.Width()

	// the height is fixed to a proportion of the EM square
	height := EMHeight * 0.5

	// adjust the baseline of the exponent scope to the height of the char
	baselineExponent := bb.UL.Y - 0.1*EMHeight
	exponent := dimsForBaseline(height, baselineExponent)

	// adjust the baseline of the indice scope just under the char
	baselineIndice := bb.LR.Y + 0.3*EMHeight
	indice := dimsForBaseline(height, baselineIndice)

	return &regularChar{grapheme: gr, exponent: newNode(exponent, xLeft), indice: newNode(indice, xLeft)}
}

func (r *regularChar) Children() []*Node { return []*Node{r.indice, r.exponent} }

// return a rect such that its baseline (top + height*EMBaselineRatio) is at baseline
func dimsForBaseline(height, baseline Fl) sy.HeightGrid {
	top := baseline - height*EMBaselineRatio
	return sy.HeightGrid{
		Ymin: top, Ymax: top + height,
		Baseline: baseline,
	}
}

func baselineFromHeight(top, bottom Fl) sy.HeightGrid {
	baseline := top + (bottom-top)*EMBaselineRatio
	return sy.HeightGrid{Ymin: top, Ymax: bottom, Baseline: baseline}
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
	numTop, numBottom := fracTop-numHeight, fracTop
	numDims := baselineFromHeight(numTop, numBottom)

	// enlarge the den height
	denHeight := EMHeight * 0.9
	denTop, denBottom := fracBottom, fracBottom+denHeight
	denDims := baselineFromHeight(denTop, denBottom)

	return &fracOperator{grapheme: gr, num: newNode(numDims, bbox.UL.X), den: newNode(denDims, bbox.UL.X)}
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
