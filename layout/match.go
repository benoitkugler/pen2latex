package layout

import (
	"fmt"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

// FindContext returns the enclosing area for the given point, usually the mouse cursor.
func (line *Line) FindContext(pos sy.Pos) sy.Context {
	node, _ := line.findNode(sy.Rect{UL: pos, LR: pos})
	return node.context(true)
}

// Insert finds the best place in [line] to insert [content],
// and updates the line.
// It also returns how the current record should be updated.
func (line *Line) Insert(rec Record, db *sy.Store) RecordAction {
	// find the correct scope
	_, last := rec.split()
	node, insertPos := line.findNode(last.BoundingBox())
	fmt.Printf("insert at index %v in node %p\n", insertPos, node)

	context := node.context(true)
	r, action, isCompound := rec.Identify(db, context)

	wholeSymbol, _, lastStroke := rec.footprints()

	if isCompound {
		// if a compound symbol is matched, simply update the previous char
		gr := grapheme{Char: r, Symbol: wholeSymbol}
		node.blocks[insertPos] = newRegularChar(gr)
	} else {
		gr := grapheme{Char: r, Symbol: sy.Footprint{Strokes: []sy.Stroke{lastStroke}}}
		// add a new block
		node.insertAt(newRegularChar(gr), insertPos)
	}

	fmt.Printf("root : %p ; tree : %v\n", &line.root, line.root)

	return action
}

func newBlock(r rune, symbol sy.Footprint) block {
	gr := grapheme{Char: r, Symbol: symbol}
	// TODO : support more operators
	switch r {
	case '_':
		return newfracOperator(gr)
	default:
		return newRegularChar(gr)
	}
}

// recursively walk the node boxes to find where to insert [glyph] :
// it should be at out.Blocks[index]
func (line *Line) findNode(glyph sy.Rect) (out *Node, index int) {
	var aux func(*Node) (*Node, int)
	aux = func(n *Node) (*Node, int) {
		// collect the extended box of each sub expression, with margin
		boxes := make([]sy.Rect, len(n.blocks))
		for i, char := range n.blocks {
			boxes[i] = blockBox(char, true)
		}
		// select the correct one ...
		index := isRectInAreas(glyph, boxes)
		// ... if index is not -1, recurse on children,
		if index != -1 {
			// select the correct scope inside the block
			childBlock := n.blocks[index]
			childNodes := childBlock.Children()
			childBoxes := make([]sy.Rect, len(childNodes))
			for i, char := range childNodes {
				childBoxes[i] = char.context(true).Box
			}
			if index := isRectInAreas(glyph, childBoxes); index != -1 { // recurse on node
				return aux(childNodes[index])
			}

			// else we are inside [childBlock], but not on one of its children
		}

		// else, we are at the correct level
		boxesWithoutMargins := make([]sy.Rect, len(n.blocks))
		for i, char := range n.blocks {
			boxesWithoutMargins[i] = blockBox(char, false)
		}
		return n, indexInsertRectBetweenArea(glyph, boxesWithoutMargins)
	}

	return aux(&line.root)
}

// isRectInAreas performs approximate matching between the given [glyph] bounding box
// and the [candidates], applying the following rules :
//   - if [glyph] is included in one candidate it returns its index
//   - if at least 60% of [glyph] area is included in a candidate, it returns its index
//   - else it returns -1
func isRectInAreas(glyph sy.Rect, candidates []sy.Rect) int {
	glyphArea := glyph.Area()
	for index, candidate := range candidates {
		if glyphArea == 0 && candidate.Contains(glyph.LR) {
			return index
		}

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
