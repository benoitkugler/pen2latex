package layout

import (
	"fmt"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

// Insert finds the best place in [line] to insert [content],
// and updates the line.
func (line *Line) Insert(rec Record, db *sy.Store) (isCompound bool) {
	// find the correct scope
	_, last := rec.split()
	node, scope, insertPos := line.FindNode(last.BoundingBox())
	fmt.Printf("enclosing box %v index : %v %p\n", scope, insertPos, node)

	// symbol := rec.InferSymbol()
	// fmt.Println("matching symbol with len", len(symbol))
	// r, _ := db.Lookup(symbol, sy.Rect{})

	// if r == 0 && symbol.IsCompound() {
	// 	// try again without the compound is the touch is spurious
	// 	_, last := rec.split()
	// 	symbol = sy.Symbol{last}
	// 	r, _ = db.Lookup(symbol, sy.Rect{})
	// }
	r, _ := rec.Identify(db)

	// if a compound symbol is matched, simply update the previous char
	// TODO:
	var symbol sy.Symbol
	if symbol.IsCompound() && line.cursor != nil {
		fmt.Println(r, string(r))
		*line.cursor = Grapheme{Char: r, Symbol: symbol}
	} else { // find the place to insert the new symbol
		// TODO:
		regChar := newChar(r, symbol)
		node.insertAt(regChar, insertPos)
		line.cursor = regChar.Content()
	}

	fmt.Printf("root : %p ; tree : %v\n", &line.root, line.root)

	return symbol.IsCompound()
}

func newChar(r rune, symbol sy.Symbol) block {
	fmt.Println("adding char", string(r))
	grapheme := Grapheme{Char: r, Symbol: symbol}
	// TODO : support more operators
	switch r {
	case '_':
		// FIXME
		return &fracOperator{Grapheme: grapheme, num: &Node{}, den: &Node{}}
	default:
		return &regularChar{Grapheme: grapheme, indice: &Node{}, exponent: &Node{}}
	}
}

// recursively walk the node boxes to find where to insert [glyph] :
// it should be at out.Blocks[index]
func (line *Line) FindNode(glyph sy.Rect) (out *Node, scope Scope, index int) {
	var aux func(*Node, Scope) (*Node, Scope, int)
	aux = func(n *Node, parentScope Scope) (*Node, Scope, int) {
		// collect the extended box of each sub expression,
		// with and without scopes
		boxes := make([]sy.Rect, len(n.Blocks))
		boxesWithoutScopes := make([]sy.Rect, len(n.Blocks))
		for i, char := range n.Blocks {
			boxes[i], boxesWithoutScopes[i] = extendedBox(char)
		}
		// select the correct one ...
		index := isRectInAreas(glyph, boxes)
		// ... if index is not -1, recurse on children,
		if index != -1 {
			// select the correct scope inside the block
			childBlock := n.Blocks[index]
			childNodes := childBlock.Children()
			childScopes := childBlock.Scopes()
			if index := isRectInAreas(glyph, childScopes); index != -1 { // recurse on node
				return aux(childNodes[index], childScopes[index])
			}

			// else we are inside [childBlock], but not on one of its children
		}

		// else, we are at the correct level
		return n, parentScope, indexInsertRectBetweenArea(glyph, boxesWithoutScopes)
	}

	return aux(&line.root, sy.EmptyRect())
}

// isRectInAreas performs approximate matching between the given [glyph] bounding box
// and the [candidates], applying the following rules :
//   - if [glyph] is included in one candidate it returns its index
//   - if at least 60% of [glyph] area is included in a candidate, it returns its index
//   - else it returns -1
func isRectInAreas(glyph sy.Rect, candidates []sy.Rect) int {
	glyphArea := glyph.Area()
	for index, candidate := range candidates {
		commonArea := candidate.Intersection(glyph).Area()
		if commonArea/glyphArea >= 0.6 {
			return index
		}
	}
	return -1
}

// indexInsertRectBetweenArea returns where to insert the given [glyph] in [candidates],
// assuming candidates are "sorted by X"
func indexInsertRectBetweenArea(glyph sy.Rect, candidates []sy.Rect) int {
	fmt.Println(glyph, candidates)
	middle := (glyph.LR.X + glyph.UL.X) / 2
	for index, candidate := range candidates {
		middleC := (candidate.LR.X + candidate.UL.X) / 2
		if middle > middleC { // keep going right
			continue
		}
		return index
	}
	return len(candidates)
}
