package symbols

import (
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

func TestSplitBezier(t *testing.T) {
	be := Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	var t0, t1 fl = 0.2, 0.4
	center := be.splitBetween(t0, t1)
	tu.Assert(t, almostEqualPos(center.P0, be.pointAt(t0)))
	tu.Assert(t, almostEqualPos(center.P3, be.pointAt(t1)))

	_ = be.String()
}
