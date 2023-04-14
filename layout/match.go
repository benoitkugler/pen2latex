package layout

import (
	"fmt"

	sy "github.com/benoitkugler/pen2latex/symbols"
)

type Context struct {
	sy.Rect
	Baseline Fl
}

// FindContext returns the enclosing area for the given point, usually the mouse cursor.
func (line *Line) FindContext(pos sy.Pos) Context {
	node, _ := line.findNode(sy.Rect{UL: pos, LR: pos})
	_, rect := node.boxes()
	// use content for X dims, but ref height
	rect.UL.Y = node.height.Ymin
	rect.LR.Y = node.height.Ymax
	return Context{Rect: rect, Baseline: node.height.Baseline}
}

// Insert finds the best place in [line] to insert [content],
// and updates the line.
// It also returns how the current record should be updated.
func (line *Line) Insert(rec Record, db *sy.Store) RecordAction {
	// find the correct scope
	_, last := rec.split()
	node, insertPos := line.findNode(last.BoundingBox())
	fmt.Printf("insert at index %v in node %p\n", insertPos, node)

	r, action, isCompound := rec.Identify(db, node.height)

	wholeSymbol, _, lastStroke := rec.footprints()

	if isCompound {
		// if a compound symbol is matched, simply update the last block
		gr := grapheme{Char: r, Symbol: wholeSymbol}
		*line.cursor = newBlock(gr)
	} else {
		gr := grapheme{Char: r, Symbol: sy.Footprint{Strokes: []sy.Stroke{lastStroke}}}
		// add a new block
		node.insertAt(newBlock(gr), insertPos)
		// .. and save the 'cursor' location
		line.cursor = &node.blocks[insertPos]
	}

	fmt.Printf("root : %p ; tree : %v\n", &line.root, line.root)

	return action
}

func newBlock(gr grapheme) block {
	// TODO : support more operators
	switch gr.Char {
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
		// collect the extended box of each sub expression
		innerBoxes, outerBoxes := make([]sy.Rect, len(n.blocks)), make([]sy.Rect, len(n.blocks))
		for i, char := range n.blocks {
			innerBoxes[i], outerBoxes[i] = blockBoxes(char)
		}
		// select the correct one ...
		index := isRectInAreas(glyph, outerBoxes)
		// ... if index is not -1, recurse on children,
		if index != -1 {
			// select the correct scope inside the block
			childBlock := n.blocks[index]
			childNodes := childBlock.Children()
			childOuterBoxes := make([]sy.Rect, len(childNodes))
			// use the parent width
			for i, char := range childNodes {
				_, childOuterBoxes[i] = char.boxes()
			}
			if index := isRectInAreas(glyph, childOuterBoxes); index != -1 { // recurse on node
				return aux(childNodes[index])
			}

			// else we are inside [childBlock], but not on one of its children
		}

		// else, we are at the correct level
		return n, indexInsertRectBetweenArea(glyph, innerBoxes)
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
