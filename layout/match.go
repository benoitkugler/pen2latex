package layout

import (
	"fmt"

	"github.com/benoitkugler/pen2latex/symbols"
)

// Insert finds the best place in [line] to insert [content],
// and updates the line.
func (line *Line) Insert(rec symbols.Record, db *symbols.SymbolStore) {
	// find the correct scope
	node, scope, insertPos := line.FindNode(rec.Shape().BoundingBox())
	fmt.Printf("enclosing box %v %v %p\n", scope, insertPos, node)

	r, preferCompound := db.Lookup(rec)

	fmt.Println(preferCompound)
	// if a compound symbol is matched, simply update the previous char
	if preferCompound && line.cursor != nil {
		*line.cursor = Grapheme{Char: r, Symbol: rec.Compound()}
	} else { // find the place to insert the new symbol
		// TODO:
		regChar := newChar(r, symbols.Symbol{rec.Shape()})
		node.insertAt(regChar, insertPos)
		line.cursor = regChar.Content()
	}

	fmt.Printf("root : %p ; tree : %v\n", &line.root, line.root)
}

func newChar(r rune, symbol symbols.Symbol) block {
	// TODO : support operators
	return &regularChar{Grapheme: Grapheme{Char: r, Symbol: symbol}, indice: &Node{}, exponent: &Node{}}
}

// recursively walk the node boxes to find where to insert [glyph] :
// it should be at out.Blocks[index]
func (line *Line) FindNode(glyph symbols.Rect) (out *Node, scope Scope, index int) {
	var aux func(*Node, Scope) (*Node, Scope, int)
	aux = func(n *Node, parentScope Scope) (*Node, Scope, int) {
		// collect the extended box of each sub expression
		boxes := make([]symbols.Rect, len(n.Blocks))
		for i, char := range n.Blocks {
			boxes[i] = extendedBox(char)
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
		return n, parentScope, indexInsertRectBetweenArea(glyph, boxes)
	}

	return aux(&line.root, symbols.EmptyRect())
}

// isRectInAreas performs approximate matching between the given [glyph] bounding box
// and the [candidates], applying the following rules :
//   - if [glyph] is included in one candidate it returns its index
//   - if at least 60% of [glyph] area is included in a candidate, it returns its index
//   - else it returns -1
func isRectInAreas(glyph symbols.Rect, candidates []symbols.Rect) int {
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
func indexInsertRectBetweenArea(glyph symbols.Rect, candidates []symbols.Rect) int {
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
