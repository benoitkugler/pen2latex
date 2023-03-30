package symbols

import (
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

func TestSplitBezier(t *testing.T) {
	be := Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	var t0, t1 Fl = 0.2, 0.4
	center := be.splitBetween(t0, t1)
	tu.Assert(t, almostEqualPos(center.P0, be.pointAt(t0)))
	tu.Assert(t, almostEqualPos(center.P3, be.pointAt(t1)))

	_ = be.String()
}

func TestIsRoughlyLinear(t *testing.T) {
	be := Bezier{Pos{}, Pos{10, 0}, Pos{20, 0}, Pos{30, 0}}
	tu.AssertEqual(t, be.IsRoughlyLinear(), true)

	be = Bezier{Pos{}, Pos{10, 30}, Pos{20, 30}, Pos{30, 0}}
	tu.AssertEqual(t, be.IsRoughlyLinear(), false)

	points := Shape{{X: 62.0, Y: 150.3}, {X: 63.0, Y: 150.3}, {X: 64.0, Y: 150.3}, {X: 66.0, Y: 150.3}, {X: 67.0, Y: 150.3}, {X: 69.0, Y: 150.3}, {X: 70.0, Y: 150.3}, {X: 72.0, Y: 150.3}, {X: 74.0, Y: 150.3}, {X: 75.0, Y: 150.3}, {X: 77.0, Y: 150.3}, {X: 78.0, Y: 150.3}, {X: 79.0, Y: 150.3}, {X: 80.0, Y: 150.3}}
	fitted := mergeSimilarCurves(fitCubicBeziers(points))
	tu.AssertEqual(t, len(fitted), 1)
	tu.AssertEqual(t, fitted[0].IsRoughlyLinear(), true)

	points = Shape{{X: 49.0, Y: 151.3}, {X: 50.0, Y: 151.3}, {X: 51.0, Y: 151.3}, {X: 53.0, Y: 150.3}, {X: 54.0, Y: 150.3}, {X: 57.0, Y: 149.3}, {X: 59.0, Y: 149.3}, {X: 62.0, Y: 148.3}, {X: 64.0, Y: 148.3}, {X: 66.0, Y: 148.3}, {X: 68.0, Y: 148.3}, {X: 70.0, Y: 148.3}, {X: 71.0, Y: 148.3}}
	fitted = mergeSimilarCurves(fitCubicBeziers(points))
	tu.AssertEqual(t, len(fitted), 1)
	tu.AssertEqual(t, fitted[0].IsRoughlyLinear(), true)

	points = Shape{{X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 56.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 57.0, Y: 52.0}, {X: 58.0, Y: 52.0}, {X: 60.0, Y: 52.0}, {X: 62.0, Y: 52.0}, {X: 64.0, Y: 52.0}, {X: 66.0, Y: 52.0}, {X: 68.0, Y: 51.0}, {X: 70.0, Y: 51.0}, {X: 72.0, Y: 51.0}, {X: 73.0, Y: 51.0}, {X: 74.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 50.0}, {X: 75.0, Y: 49.0}, {X: 75.0, Y: 49.0}, {X: 74.0, Y: 50.0}}
	fitted = mergeSimilarCurves(fitCubicBeziers(points))
	tu.AssertEqual(t, len(fitted), 1)
	tu.AssertEqual(t, fitted[0].IsRoughlyLinear(), true)
}

func TestIntersects(t *testing.T) {
	vertical := Shape{{X: 58.0, Y: 139.3}, {X: 58.0, Y: 140.3}, {X: 58.0, Y: 141.3}, {X: 58.0, Y: 142.3}, {X: 58.0, Y: 143.3}, {X: 58.0, Y: 144.3}, {X: 58.0, Y: 146.3}, {X: 58.0, Y: 147.3}, {X: 59.0, Y: 149.3}, {X: 59.0, Y: 150.3}, {X: 59.0, Y: 152.3}, {X: 59.0, Y: 154.3}, {X: 59.0, Y: 155.3}, {X: 60.0, Y: 157.3}, {X: 60.0, Y: 158.3}, {X: 60.0, Y: 159.3}, {X: 60.0, Y: 160.3}, {X: 60.0, Y: 162.3}, {X: 61.0, Y: 163.3}, {X: 61.0, Y: 164.3}, {X: 61.0, Y: 165.3}, {X: 61.0, Y: 167.3}, {X: 61.0, Y: 168.3}, {X: 61.0, Y: 169.3}, {X: 61.0, Y: 170.3}, {X: 61.0, Y: 171.3}, {X: 61.0, Y: 172.3}, {X: 61.0, Y: 174.3}, {X: 61.0, Y: 175.3}, {X: 61.0, Y: 176.3}, {X: 61.0, Y: 177.3}, {X: 61.0, Y: 178.3}, {X: 61.0, Y: 179.3}, {X: 61.0, Y: 180.3}, {X: 61.0, Y: 181.3}, {X: 61.0, Y: 183.3}, {X: 61.0, Y: 184.3}, {X: 61.0, Y: 185.3}, {X: 61.0, Y: 187.3}, {X: 61.0, Y: 188.3}, {X: 61.0, Y: 190.3}, {X: 61.0, Y: 191.3}, {X: 61.0, Y: 192.3}, {X: 62.0, Y: 193.3}, {X: 62.0, Y: 194.3}, {X: 63.0, Y: 195.3}, {X: 64.0, Y: 195.3}, {X: 65.0, Y: 195.3}, {X: 66.0, Y: 195.3}, {X: 67.0, Y: 195.3}, {X: 68.0, Y: 195.3}, {X: 69.0, Y: 195.3}, {X: 70.0, Y: 195.3}, {X: 70.0, Y: 194.3}, {X: 71.0, Y: 194.3}, {X: 72.0, Y: 193.3}, {X: 72.0, Y: 192.3}}
	horizontal := Shape{{X: 49.0, Y: 151.3}, {X: 50.0, Y: 151.3}, {X: 51.0, Y: 151.3}, {X: 53.0, Y: 150.3}, {X: 54.0, Y: 150.3}, {X: 57.0, Y: 149.3}, {X: 59.0, Y: 149.3}, {X: 62.0, Y: 148.3}, {X: 64.0, Y: 148.3}, {X: 66.0, Y: 148.3}, {X: 68.0, Y: 148.3}, {X: 70.0, Y: 148.3}, {X: 71.0, Y: 148.3}}

	v := mergeSimilarCurves(fitCubicBeziers(vertical))
	h := mergeSimilarCurves(fitCubicBeziers(horizontal))
	tu.AssertEqual(t, len(h), 1)
	tu.AssertEqual(t, h[0].IsRoughlyLinear(), true)
	start, end := h[0].P0, h[0].P3
	tu.AssertEqual(t, v[0].IntersectsSegment(start, end), true)

	vertical = Shape{{X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 38.0}, {X: 56.0, Y: 39.0}, {X: 56.0, Y: 39.0}, {X: 56.0, Y: 39.0}, {X: 56.0, Y: 40.0}, {X: 56.0, Y: 41.0}, {X: 56.0, Y: 42.0}, {X: 56.0, Y: 43.0}, {X: 56.0, Y: 45.0}, {X: 55.0, Y: 46.0}, {X: 55.0, Y: 48.0}, {X: 55.0, Y: 50.0}, {X: 55.0, Y: 52.0}, {X: 54.0, Y: 54.0}, {X: 54.0, Y: 57.0}, {X: 54.0, Y: 59.0}, {X: 54.0, Y: 62.0}, {X: 53.0, Y: 65.0}, {X: 53.0, Y: 68.0}, {X: 53.0, Y: 71.0}, {X: 53.0, Y: 74.0}, {X: 53.0, Y: 77.0}, {X: 53.0, Y: 81.0}, {X: 54.0, Y: 83.0}, {X: 54.0, Y: 86.0}, {X: 54.0, Y: 89.0}, {X: 55.0, Y: 91.0}, {X: 56.0, Y: 93.0}, {X: 56.0, Y: 95.0}, {X: 57.0, Y: 96.0}, {X: 58.0, Y: 97.0}, {X: 58.0, Y: 98.0}, {X: 59.0, Y: 99.0}, {X: 59.0, Y: 99.0}, {X: 60.0, Y: 100.0}, {X: 60.0, Y: 100.0}, {X: 61.0, Y: 100.0}, {X: 62.0, Y: 99.0}, {X: 63.0, Y: 98.0}, {X: 64.0, Y: 96.0}, {X: 66.0, Y: 94.0}}
	horizontal = Shape{{X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 48.0, Y: 50.0}, {X: 49.0, Y: 50.0}, {X: 49.0, Y: 50.0}, {X: 51.0, Y: 50.0}, {X: 52.0, Y: 50.0}, {X: 53.0, Y: 50.0}, {X: 55.0, Y: 49.0}, {X: 56.0, Y: 49.0}, {X: 58.0, Y: 49.0}, {X: 59.0, Y: 49.0}, {X: 60.0, Y: 49.0}, {X: 60.0, Y: 49.0}, {X: 61.0, Y: 49.0}, {X: 61.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 62.0, Y: 49.0}, {X: 61.0, Y: 50.0}}
	v = mergeSimilarCurves(fitCubicBeziers(vertical))
	h = mergeSimilarCurves(fitCubicBeziers(horizontal))
	tu.AssertEqual(t, len(h), 1)
	tu.AssertEqual(t, h[0].IsRoughlyLinear(), true)
	start, end = h[0].P0, h[0].P3
	tu.AssertEqual(t, v[0].IntersectsSegment(start, end), true)
}
