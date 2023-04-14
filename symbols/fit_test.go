package symbols

import (
	"math"
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

func TestPathIndices(t *testing.T) {
	points := []Pos{
		{0, 0}, {1, 1}, {1, 1}, {1, 8},
	}
	indices := pathLengthIndices(points)
	tu.Assert(t, indices[0] == 0 && indices[len(indices)-1] == 1)
}

func generateCircle(center Pos, radius Fl, nbPoints int) Shape {
	var out Shape
	for i := 0; i < nbPoints; i++ {
		theta := 2 * math.Pi * float64(i) / float64(nbPoints)
		out = append(out, center.Add(Pos{
			X: float32(math.Cos(theta)),
			Y: float32(math.Sin(theta)),
		}.ScaleTo(radius)))
	}
	return out
}

func TestFitBezier(t *testing.T) {
	origin := Bezier{Pos{}, Pos{30, 40}, Pos{50, -40}, Pos{60, 0}}
	points := origin.toPoints()
	printShape(t, points, "bezier_cube1_origin.png")

	fitted, _ := fitCubicBezier(points)
	printShape(t, fitted.toPoints(), "bezier_cube1_fitted.png")

	origin = Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	points = origin.toPoints()
	printShape(t, points, "bezier_cube2_origin.png")

	fitted, _ = fitCubicBezier(points)
	printShape(t, fitted.toPoints(), "bezier_cube2_fitted.png")

	// linear shape
	origin = Bezier{Pos{}, Pos{30, 30}, Pos{40, 40}, Pos{60, 60}}
	points = origin.toPoints()
	printShape(t, points, "bezier_cube3_origin.png")

	fitted, _ = fitCubicBezier(points)
	printShape(t, fitted.toPoints(), "bezier_cube3_fitted.png")

	_, errLine := fitSegment(points)
	tu.Assert(t, errLine == 0)

	// data from a circle, not a bezier !
	points = generateCircle(Pos{30, 30}, 20, 20)
	fitted, _ = fitCubicBezier(points)
	printShape(t, fitted.toPoints(), "bezier_cube4_fitted.png")
}

func almostEqual(u, v Fl) bool     { return abs(u-v) < 1e-3 }
func almostEqualPos(u, v Pos) bool { return almostEqual(u.X, v.X) && almostEqual(u.Y, v.Y) }

func TestFitBezierArtefact(t *testing.T) {
	// artifact at the end
	points := Shape{
		{X: 74.0, Y: 192.3}, {X: 74.0, Y: 191.3}, {X: 74.0, Y: 190.3}, {X: 74.0, Y: 189.3}, {X: 75.0, Y: 187.3}, {X: 75.0, Y: 185.3}, {X: 76.0, Y: 183.3}, {X: 76.0, Y: 180.3}, {X: 77.0, Y: 177.3}, {X: 77.0, Y: 174.3}, {X: 78.0, Y: 170.3}, {X: 78.0, Y: 167.3}, {X: 78.0, Y: 163.3}, {X: 78.0, Y: 160.3}, {X: 79.0, Y: 157.3}, {X: 79.0, Y: 154.3}, {X: 79.0, Y: 152.3}, {X: 79.0, Y: 151.3}, {X: 79.0, Y: 149.3}, {X: 79.0, Y: 148.3}, {X: 79.0, Y: 147.3}, {X: 79.0, Y: 146.3}, {X: 80.0, Y: 147.3}, {X: 81.0, Y: 148.3}, {X: 82.0, Y: 150.3}, {X: 83.0, Y: 152.3}, {X: 85.0, Y: 154.3}, {X: 87.0, Y: 156.3}, {X: 90.0, Y: 158.3}, {X: 93.0, Y: 160.3}, {X: 96.0, Y: 161.3}, {X: 98.0, Y: 162.3}, {X: 102.0, Y: 162.3}, {X: 105.0, Y: 161.3}, {X: 108.0, Y: 161.3}, {X: 111.0, Y: 159.3}, {X: 115.0, Y: 158.3}, {X: 118.0, Y: 156.3}, {X: 120.0, Y: 154.3}, {X: 122.0, Y: 151.3}, {X: 124.0, Y: 149.3}, {X: 126.0, Y: 148.3}, {X: 127.0, Y: 146.3}, {X: 127.0, Y: 145.3}, {X: 127.0, Y: 146.3}, {X: 127.0, Y: 147.3}, {X: 127.0, Y: 150.3}, {X: 126.0, Y: 153.3}, {X: 126.0, Y: 156.3}, {X: 125.0, Y: 160.3}, {X: 124.0, Y: 164.3}, {X: 124.0, Y: 168.3}, {X: 123.0, Y: 171.3}, {X: 123.0, Y: 175.3}, {X: 123.0, Y: 178.3}, {X: 123.0, Y: 181.3}, {X: 123.0, Y: 183.3}, {X: 123.0, Y: 186.3}, {X: 123.0, Y: 187.3}, {X: 123.0, Y: 189.3}, {X: 123.0, Y: 190.3}, {X: 124.0, Y: 191.3}, {X: 124.0, Y: 192.3}, {X: 124.0, Y: 193.3}, {X: 124.0, Y: 194.3}, {X: 125.0, Y: 194.3}, {X: 125.0, Y: 195.3}, {X: 125.0, Y: 196.3}, {X: 126.0, Y: 196.3}, {X: 126.0, Y: 195.3},
	}

	tangent := computeEndTangent(removeSideArtifacts(points))
	tu.Assert(t, tangent.Y < 0)
}

func TestFitBeziers(t *testing.T) {
	points := Bezier{Pos{}, Pos{30, 40}, Pos{50, -40}, Pos{60, 0}}.toPoints()
	fitteds := fitCubicBeziers(points)
	tu.Assert(t, len(fitteds) == 1)

	points = Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}.toPoints()
	fitteds = fitCubicBeziers(points)
	tu.Assert(t, len(fitteds) == 1)

	// linear shape
	points = Bezier{Pos{}, Pos{30, 30}, Pos{40, 40}, Pos{60, 60}}.toPoints()
	fitteds = fitCubicBeziers(points)
	tu.Assert(t, len(fitteds) == 1)

	// circular shape
	points = generateCircle(Pos{30, 30}, 20, 20)
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(fitteds), 3)

	// paren shape
	points = Shape{
		{X: 98.0, Y: 9.3}, {X: 99.0, Y: 10.3}, {X: 100.0, Y: 11.3}, {X: 101.0, Y: 12.3}, {X: 102.0, Y: 13.3}, {X: 102.0, Y: 14.3}, {X: 103.0, Y: 15.3}, {X: 103.0, Y: 16.3}, {X: 104.0, Y: 17.3}, {X: 104.0, Y: 18.3}, {X: 104.0, Y: 19.3}, {X: 104.0, Y: 20.3}, {X: 105.0, Y: 21.3}, {X: 105.0, Y: 22.3}, {X: 105.0, Y: 23.3}, {X: 105.0, Y: 24.3}, {X: 105.0, Y: 25.3}, {X: 105.0, Y: 27.3}, {X: 105.0, Y: 29.3}, {X: 105.0, Y: 30.3}, {X: 105.0, Y: 32.3}, {X: 105.0, Y: 34.3}, {X: 105.0, Y: 35.3}, {X: 105.0, Y: 36.3}, {X: 105.0, Y: 38.3}, {X: 104.0, Y: 38.3}, {X: 104.0, Y: 39.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.Assert(t, len(fitteds) == 1)

	// Sigma
	points = Shape{
		{X: 107.0, Y: 90.3}, {X: 105.0, Y: 90.3}, {X: 104.0, Y: 91.3}, {X: 102.0, Y: 91.3}, {X: 99.0, Y: 91.3}, {X: 96.0, Y: 92.3}, {X: 93.0, Y: 92.3}, {X: 90.0, Y: 93.3}, {X: 86.0, Y: 93.3}, {X: 82.0, Y: 94.3}, {X: 78.0, Y: 94.3}, {X: 74.0, Y: 95.3}, {X: 70.0, Y: 95.3}, {X: 66.0, Y: 96.3}, {X: 63.0, Y: 96.3}, {X: 60.0, Y: 97.3}, {X: 58.0, Y: 97.3}, {X: 56.0, Y: 97.3}, {X: 55.0, Y: 97.3}, {X: 54.0, Y: 97.3}, {X: 53.0, Y: 97.3}, {X: 53.0, Y: 98.3}, {X: 54.0, Y: 99.3}, {X: 55.0, Y: 100.3}, {X: 57.0, Y: 101.3}, {X: 59.0, Y: 103.3}, {X: 61.0, Y: 104.3}, {X: 63.0, Y: 105.3}, {X: 65.0, Y: 107.3}, {X: 68.0, Y: 109.3}, {X: 71.0, Y: 111.3}, {X: 74.0, Y: 112.3}, {X: 77.0, Y: 114.3}, {X: 80.0, Y: 116.3}, {X: 82.0, Y: 117.3}, {X: 85.0, Y: 119.3}, {X: 87.0, Y: 120.3}, {X: 89.0, Y: 121.3}, {X: 90.0, Y: 122.3}, {X: 92.0, Y: 123.3}, {X: 93.0, Y: 124.3}, {X: 93.0, Y: 125.3}, {X: 94.0, Y: 125.3}, {X: 94.0, Y: 126.3}, {X: 95.0, Y: 126.3}, {X: 94.0, Y: 126.3}, {X: 94.0, Y: 127.3}, {X: 93.0, Y: 128.3}, {X: 92.0, Y: 129.3}, {X: 91.0, Y: 130.3}, {X: 90.0, Y: 132.3}, {X: 89.0, Y: 133.3}, {X: 88.0, Y: 135.3}, {X: 86.0, Y: 138.3}, {X: 84.0, Y: 140.3}, {X: 82.0, Y: 143.3}, {X: 80.0, Y: 146.3}, {X: 78.0, Y: 149.3}, {X: 75.0, Y: 152.3}, {X: 73.0, Y: 154.3}, {X: 72.0, Y: 156.3}, {X: 70.0, Y: 158.3}, {X: 69.0, Y: 159.3}, {X: 68.0, Y: 160.3}, {X: 68.0, Y: 161.3}, {X: 67.0, Y: 162.3}, {X: 68.0, Y: 162.3}, {X: 69.0, Y: 162.3}, {X: 71.0, Y: 162.3}, {X: 73.0, Y: 162.3}, {X: 76.0, Y: 162.3}, {X: 80.0, Y: 162.3}, {X: 84.0, Y: 162.3}, {X: 88.0, Y: 162.3}, {X: 93.0, Y: 161.3}, {X: 98.0, Y: 161.3}, {X: 104.0, Y: 160.3}, {X: 108.0, Y: 159.3}, {X: 112.0, Y: 159.3}, {X: 115.0, Y: 158.3}, {X: 118.0, Y: 158.3}, {X: 121.0, Y: 157.3}, {X: 123.0, Y: 156.3}, {X: 125.0, Y: 156.3}, {X: 126.0, Y: 156.3}, {X: 127.0, Y: 155.3}, {X: 127.0, Y: 156.3}, {X: 126.0, Y: 157.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(fitteds), 4)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 4)
	points = Shape{
		{X: 108.0, Y: 101.3}, {X: 107.0, Y: 101.3}, {X: 106.0, Y: 101.3}, {X: 105.0, Y: 101.3}, {X: 104.0, Y: 100.3}, {X: 103.0, Y: 100.3}, {X: 101.0, Y: 100.3}, {X: 100.0, Y: 100.3}, {X: 98.0, Y: 100.3}, {X: 96.0, Y: 100.3}, {X: 94.0, Y: 100.3}, {X: 92.0, Y: 100.3}, {X: 90.0, Y: 100.3}, {X: 87.0, Y: 101.3}, {X: 85.0, Y: 101.3}, {X: 82.0, Y: 101.3}, {X: 80.0, Y: 101.3}, {X: 77.0, Y: 102.3}, {X: 74.0, Y: 102.3}, {X: 72.0, Y: 102.3}, {X: 70.0, Y: 102.3}, {X: 69.0, Y: 102.3}, {X: 68.0, Y: 102.3}, {X: 67.0, Y: 102.3}, {X: 67.0, Y: 103.3}, {X: 68.0, Y: 103.3}, {X: 68.0, Y: 104.3}, {X: 69.0, Y: 105.3}, {X: 71.0, Y: 106.3}, {X: 72.0, Y: 108.3}, {X: 74.0, Y: 109.3}, {X: 76.0, Y: 111.3}, {X: 78.0, Y: 113.3}, {X: 80.0, Y: 114.3}, {X: 82.0, Y: 116.3}, {X: 83.0, Y: 117.3}, {X: 85.0, Y: 118.3}, {X: 86.0, Y: 120.3}, {X: 88.0, Y: 121.3}, {X: 89.0, Y: 122.3}, {X: 90.0, Y: 123.3}, {X: 91.0, Y: 123.3}, {X: 91.0, Y: 124.3}, {X: 92.0, Y: 124.3}, {X: 92.0, Y: 125.3}, {X: 93.0, Y: 125.3}, {X: 93.0, Y: 126.3}, {X: 93.0, Y: 127.3}, {X: 93.0, Y: 128.3}, {X: 94.0, Y: 129.3}, {X: 93.0, Y: 130.3}, {X: 93.0, Y: 131.3}, {X: 92.0, Y: 132.3}, {X: 91.0, Y: 134.3}, {X: 89.0, Y: 137.3}, {X: 87.0, Y: 140.3}, {X: 85.0, Y: 142.3}, {X: 82.0, Y: 145.3}, {X: 80.0, Y: 148.3}, {X: 77.0, Y: 150.3}, {X: 75.0, Y: 152.3}, {X: 74.0, Y: 154.3}, {X: 72.0, Y: 155.3}, {X: 71.0, Y: 157.3}, {X: 70.0, Y: 158.3}, {X: 70.0, Y: 159.3}, {X: 70.0, Y: 160.3}, {X: 71.0, Y: 161.3}, {X: 73.0, Y: 161.3}, {X: 75.0, Y: 161.3}, {X: 78.0, Y: 162.3}, {X: 82.0, Y: 162.3}, {X: 86.0, Y: 162.3}, {X: 90.0, Y: 162.3}, {X: 94.0, Y: 162.3}, {X: 99.0, Y: 161.3}, {X: 103.0, Y: 161.3}, {X: 107.0, Y: 161.3}, {X: 111.0, Y: 161.3}, {X: 114.0, Y: 161.3}, {X: 116.0, Y: 161.3}, {X: 118.0, Y: 161.3}, {X: 119.0, Y: 161.3}, {X: 118.0, Y: 161.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 4)

	// e
	points = Shape{
		{X: 94.0, Y: 104.3}, {X: 95.0, Y: 105.3}, {X: 96.0, Y: 106.0}, {X: 97.3, Y: 106.6}, {X: 98.7, Y: 107.0}, {X: 100.3, Y: 107.6}, {X: 102.0, Y: 108.0}, {X: 104.0, Y: 108.3}, {X: 105.7, Y: 108.3}, {X: 107.3, Y: 108.3}, {X: 109.0, Y: 108.3}, {X: 111.0, Y: 108.3}, {X: 112.7, Y: 108.3}, {X: 114.0, Y: 108.0}, {X: 115.3, Y: 107.3}, {X: 116.7, Y: 106.6}, {X: 118.0, Y: 106.0}, {X: 119.0, Y: 105.3}, {X: 120.0, Y: 104.3}, {X: 120.7, Y: 103.3}, {X: 121.3, Y: 102.0}, {X: 122.0, Y: 100.6}, {X: 122.7, Y: 99.3}, {X: 123.3, Y: 98.0}, {X: 123.7, Y: 96.6}, {X: 124.0, Y: 95.0}, {X: 124.0, Y: 93.6}, {X: 124.3, Y: 92.3}, {X: 124.3, Y: 91.3}, {X: 124.3, Y: 90.3}, {X: 124.0, Y: 89.3}, {X: 123.7, Y: 88.3}, {X: 123.0, Y: 87.3}, {X: 122.0, Y: 86.3}, {X: 121.0, Y: 85.3}, {X: 120.0, Y: 84.3}, {X: 119.0, Y: 83.3}, {X: 117.7, Y: 82.6}, {X: 116.0, Y: 82.0}, {X: 114.3, Y: 81.6}, {X: 112.7, Y: 81.3}, {X: 111.3, Y: 81.3}, {X: 109.7, Y: 81.0}, {X: 108.0, Y: 81.0}, {X: 106.0, Y: 81.0}, {X: 104.3, Y: 81.6}, {X: 102.7, Y: 82.0}, {X: 101.3, Y: 82.6}, {X: 99.7, Y: 83.3}, {X: 98.3, Y: 84.6}, {X: 96.7, Y: 86.0}, {X: 95.3, Y: 87.6}, {X: 94.0, Y: 89.0}, {X: 93.0, Y: 90.6}, {X: 92.0, Y: 92.3}, {X: 91.0, Y: 94.6}, {X: 90.0, Y: 97.0}, {X: 89.3, Y: 99.3}, {X: 89.0, Y: 101.6}, {X: 88.7, Y: 104.3}, {X: 88.3, Y: 107.3}, {X: 88.0, Y: 110.3}, {X: 88.0, Y: 113.3}, {X: 88.3, Y: 116.6}, {X: 89.0, Y: 120.0}, {X: 90.0, Y: 123.3}, {X: 91.0, Y: 126.0}, {X: 92.3, Y: 128.6}, {X: 93.7, Y: 131.0}, {X: 95.3, Y: 133.3}, {X: 97.3, Y: 135.3}, {X: 100.0, Y: 137.0}, {X: 102.7, Y: 138.3}, {X: 105.3, Y: 139.3}, {X: 108.3, Y: 140.0}, {X: 111.7, Y: 140.3}, {X: 115.3, Y: 140.0}, {X: 118.7, Y: 139.3}, {X: 122.0, Y: 138.3}, {X: 125.0, Y: 137.0}, {X: 128.0, Y: 135.6}, {X: 130.7, Y: 134.0}, {X: 132.7, Y: 132.6}, {X: 134.0, Y: 131.3}, {X: 135.0, Y: 130.6}, {X: 135.7, Y: 130.0}, {X: 136.0, Y: 129.3}, {X: 136.0, Y: 128.3}, {X: 136.0, Y: 128.0}, {X: 136.0, Y: 128.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(fitteds), 3)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 2)

	// 2
	points = Shape{
		{X: 18.00, Y: 10.30}, {X: 18.00, Y: 9.30}, {X: 18.00, Y: 8.30}, {X: 18.00, Y: 7.30}, {X: 18.00, Y: 6.30}, {X: 18.00, Y: 5.30}, {X: 19.00, Y: 4.30}, {X: 19.00, Y: 5.30}, {X: 20.00, Y: 5.30}, {X: 21.00, Y: 7.30}, {X: 22.00, Y: 8.30}, {X: 22.00, Y: 11.30}, {X: 23.00, Y: 13.30}, {X: 23.00, Y: 16.30}, {X: 23.00, Y: 19.30}, {X: 23.00, Y: 21.30}, {X: 22.00, Y: 23.30}, {X: 21.00, Y: 25.30}, {X: 20.00, Y: 27.30}, {X: 20.00, Y: 29.30}, {X: 19.00, Y: 30.30}, {X: 18.00, Y: 30.30}, {X: 17.00, Y: 31.30}, {X: 16.00, Y: 31.30}, {X: 17.00, Y: 30.30}, {X: 18.00, Y: 30.30}, {X: 19.00, Y: 30.30}, {X: 20.00, Y: 31.30}, {X: 21.00, Y: 31.30}, {X: 22.00, Y: 32.30}, {X: 23.00, Y: 32.30}, {X: 24.00, Y: 33.30}, {X: 25.00, Y: 33.30}, {X: 26.00, Y: 32.30}, {X: 26.00, Y: 31.30}, {X: 26.00, Y: 30.30}, {X: 26.00, Y: 29.30}, {X: 26.00, Y: 28.30}, {X: 26.00, Y: 27.30}, {X: 26.00, Y: 28.30},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(fitteds), 2)

	// 3
	points = Shape{
		{X: 92.0, Y: 153.3}, {X: 91.0, Y: 153.3}, {X: 92.0, Y: 152.3}, {X: 93.0, Y: 151.3}, {X: 95.0, Y: 150.3}, {X: 96.0, Y: 149.3}, {X: 98.0, Y: 149.3}, {X: 100.0, Y: 148.3}, {X: 101.0, Y: 148.3}, {X: 103.0, Y: 148.3}, {X: 105.0, Y: 148.3}, {X: 106.0, Y: 149.3}, {X: 107.0, Y: 151.3}, {X: 108.0, Y: 152.3}, {X: 108.0, Y: 154.3}, {X: 108.0, Y: 155.3}, {X: 107.0, Y: 157.3}, {X: 106.0, Y: 159.3}, {X: 105.0, Y: 161.3}, {X: 104.0, Y: 162.3}, {X: 102.0, Y: 163.3}, {X: 101.0, Y: 164.3}, {X: 100.0, Y: 165.3}, {X: 99.0, Y: 166.3}, {X: 98.0, Y: 166.3}, {X: 98.0, Y: 167.3}, {X: 97.0, Y: 167.3}, {X: 98.0, Y: 167.3}, {X: 98.0, Y: 168.3}, {X: 99.0, Y: 168.3}, {X: 100.0, Y: 169.3}, {X: 101.0, Y: 169.3}, {X: 103.0, Y: 170.3}, {X: 104.0, Y: 171.3}, {X: 105.0, Y: 173.3}, {X: 106.0, Y: 174.3}, {X: 107.0, Y: 176.3}, {X: 107.0, Y: 178.3}, {X: 107.0, Y: 180.3}, {X: 107.0, Y: 182.3}, {X: 106.0, Y: 183.3}, {X: 105.0, Y: 185.3}, {X: 103.0, Y: 186.3}, {X: 102.0, Y: 187.3}, {X: 100.0, Y: 188.3}, {X: 98.0, Y: 189.3}, {X: 97.0, Y: 189.3}, {X: 95.0, Y: 189.3}, {X: 94.0, Y: 188.3}, {X: 92.0, Y: 188.3}, {X: 91.0, Y: 187.3}, {X: 90.0, Y: 186.3}, {X: 90.0, Y: 185.3}, {X: 89.0, Y: 185.3}, {X: 89.0, Y: 184.3}, {X: 89.0, Y: 183.3}, {X: 89.0, Y: 182.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(fitteds), 4) // two splits
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 2)

	// S
	points = Shape{{X: 113.0, Y: 4.3}, {X: 112.0, Y: 4.3}, {X: 111.0, Y: 4.3}, {X: 110.0, Y: 4.3}, {X: 109.0, Y: 5.3}, {X: 108.0, Y: 5.3}, {X: 107.0, Y: 5.3}, {X: 106.0, Y: 6.3}, {X: 105.0, Y: 7.3}, {X: 104.0, Y: 8.3}, {X: 103.0, Y: 9.3}, {X: 103.0, Y: 10.3}, {X: 102.0, Y: 11.3}, {X: 102.0, Y: 12.3}, {X: 102.0, Y: 13.3}, {X: 101.0, Y: 14.3}, {X: 101.0, Y: 15.3}, {X: 101.0, Y: 16.3}, {X: 102.0, Y: 16.3}, {X: 103.0, Y: 17.3}, {X: 104.0, Y: 17.3}, {X: 105.0, Y: 17.3}, {X: 106.0, Y: 17.3}, {X: 107.0, Y: 17.3}, {X: 108.0, Y: 17.3}, {X: 109.0, Y: 17.3}, {X: 110.0, Y: 17.3}, {X: 111.0, Y: 17.3}, {X: 112.0, Y: 17.3}, {X: 113.0, Y: 17.3}, {X: 114.0, Y: 17.3}, {X: 115.0, Y: 17.3}, {X: 116.0, Y: 18.3}, {X: 117.0, Y: 18.3}, {X: 117.0, Y: 19.3}, {X: 117.0, Y: 20.3}, {X: 117.0, Y: 21.3}, {X: 117.0, Y: 22.3}, {X: 117.0, Y: 23.3}, {X: 116.0, Y: 24.3}, {X: 115.0, Y: 24.3}, {X: 115.0, Y: 25.3}, {X: 114.0, Y: 25.3}, {X: 113.0, Y: 26.3}, {X: 112.0, Y: 26.3}, {X: 111.0, Y: 26.3}, {X: 110.0, Y: 27.3}, {X: 108.0, Y: 27.3}, {X: 107.0, Y: 27.3}, {X: 106.0, Y: 27.3}, {X: 105.0, Y: 27.3}, {X: 104.0, Y: 27.3}, {X: 103.0, Y: 27.3}}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 3)

	// a
	points = Shape{
		{X: 24.00, Y: 18.30}, {X: 23.00, Y: 17.30}, {X: 22.00, Y: 17.30}, {X: 21.00, Y: 17.30}, {X: 20.00, Y: 17.30}, {X: 19.00, Y: 17.30}, {X: 18.00, Y: 17.30}, {X: 18.00, Y: 18.30}, {X: 17.00, Y: 18.30}, {X: 17.00, Y: 20.30}, {X: 16.00, Y: 21.30}, {X: 16.00, Y: 22.30}, {X: 16.00, Y: 24.30}, {X: 16.00, Y: 25.30}, {X: 17.00, Y: 27.30}, {X: 17.00, Y: 28.30}, {X: 18.00, Y: 29.30}, {X: 19.00, Y: 29.30}, {X: 20.00, Y: 29.30}, {X: 21.00, Y: 29.30}, {X: 22.00, Y: 29.30}, {X: 23.00, Y: 28.30}, {X: 23.00, Y: 26.30}, {X: 24.00, Y: 25.30}, {X: 24.00, Y: 23.30}, {X: 24.00, Y: 21.30}, {X: 24.00, Y: 19.30}, {X: 24.00, Y: 17.30}, {X: 24.00, Y: 16.30}, {X: 24.00, Y: 15.30}, {X: 23.00, Y: 15.30}, {X: 23.00, Y: 16.30}, {X: 23.00, Y: 17.30}, {X: 23.00, Y: 19.30}, {X: 23.00, Y: 21.30}, {X: 24.00, Y: 23.30}, {X: 25.00, Y: 24.30}, {X: 26.00, Y: 26.30}, {X: 27.00, Y: 27.30}, {X: 28.00, Y: 28.30}, {X: 29.00, Y: 28.30}, {X: 30.00, Y: 29.30}, {X: 31.00, Y: 29.30}, {X: 32.00, Y: 29.30},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 3)

	// R
	points = Shape{
		{X: 49.0, Y: 108.3}, {X: 49.0, Y: 109.3}, {X: 50.0, Y: 110.3}, {X: 50.0, Y: 112.3}, {X: 50.0, Y: 114.3}, {X: 50.0, Y: 116.3}, {X: 50.0, Y: 119.3}, {X: 51.0, Y: 122.3}, {X: 51.0, Y: 126.3}, {X: 51.0, Y: 129.3}, {X: 52.0, Y: 133.3}, {X: 52.0, Y: 137.3}, {X: 52.0, Y: 141.3}, {X: 52.0, Y: 144.3}, {X: 53.0, Y: 147.3}, {X: 53.0, Y: 150.3}, {X: 53.0, Y: 152.3}, {X: 53.0, Y: 154.3}, {X: 53.0, Y: 156.3}, {X: 53.0, Y: 157.3}, {X: 53.0, Y: 156.3}, {X: 53.0, Y: 155.3}, {X: 52.0, Y: 154.3}, {X: 52.0, Y: 152.3}, {X: 52.0, Y: 151.3}, {X: 51.0, Y: 149.3}, {X: 51.0, Y: 147.3}, {X: 50.0, Y: 145.3}, {X: 50.0, Y: 143.3}, {X: 49.0, Y: 140.3}, {X: 49.0, Y: 138.3}, {X: 48.0, Y: 135.3}, {X: 48.0, Y: 133.3}, {X: 47.0, Y: 130.3}, {X: 47.0, Y: 128.3}, {X: 46.0, Y: 125.3}, {X: 46.0, Y: 123.3}, {X: 46.0, Y: 120.3}, {X: 46.0, Y: 118.3}, {X: 45.0, Y: 116.3}, {X: 46.0, Y: 114.3}, {X: 46.0, Y: 112.3}, {X: 47.0, Y: 109.3}, {X: 48.0, Y: 107.3}, {X: 49.0, Y: 105.3}, {X: 50.0, Y: 102.3}, {X: 52.0, Y: 100.3}, {X: 53.0, Y: 98.3}, {X: 55.0, Y: 97.3}, {X: 56.0, Y: 96.3}, {X: 58.0, Y: 95.3}, {X: 59.0, Y: 95.3}, {X: 61.0, Y: 95.3}, {X: 62.0, Y: 96.3}, {X: 64.0, Y: 97.3}, {X: 66.0, Y: 98.3}, {X: 67.0, Y: 100.3}, {X: 68.0, Y: 103.3}, {X: 69.0, Y: 105.3}, {X: 70.0, Y: 108.3}, {X: 70.0, Y: 111.3}, {X: 70.0, Y: 113.3}, {X: 69.0, Y: 116.3}, {X: 68.0, Y: 118.3}, {X: 67.0, Y: 120.3}, {X: 65.0, Y: 122.3}, {X: 64.0, Y: 124.3}, {X: 62.0, Y: 125.3}, {X: 61.0, Y: 126.3}, {X: 59.0, Y: 127.3}, {X: 58.0, Y: 127.3}, {X: 57.0, Y: 127.3}, {X: 56.0, Y: 127.3}, {X: 55.0, Y: 127.3}, {X: 55.0, Y: 126.3}, {X: 56.0, Y: 126.3}, {X: 57.0, Y: 127.3}, {X: 58.0, Y: 127.3}, {X: 61.0, Y: 129.3}, {X: 64.0, Y: 130.3}, {X: 67.0, Y: 133.3}, {X: 71.0, Y: 135.3}, {X: 75.0, Y: 138.3}, {X: 79.0, Y: 141.3}, {X: 84.0, Y: 144.3}, {X: 87.0, Y: 146.3}, {X: 91.0, Y: 149.3}, {X: 94.0, Y: 151.3}, {X: 96.0, Y: 153.3}, {X: 98.0, Y: 155.3}, {X: 100.0, Y: 156.3}, {X: 101.0, Y: 158.3}, {X: 101.0, Y: 159.3}, {X: 102.0, Y: 159.3}, {X: 102.0, Y: 160.3},
	}
	fitteds = fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 4)
}

func TestR(t *testing.T) {
	points := Shape{{X: 88.0, Y: 152.3}, {X: 88.0, Y: 151.3}, {X: 88.0, Y: 150.3}, {X: 88.0, Y: 149.3}, {X: 89.0, Y: 148.3}, {X: 90.0, Y: 147.3}, {X: 90.0, Y: 146.3}, {X: 91.0, Y: 145.3}, {X: 92.0, Y: 144.3}, {X: 93.0, Y: 144.3}, {X: 94.0, Y: 143.3}, {X: 95.0, Y: 143.3}, {X: 96.0, Y: 143.3}, {X: 97.0, Y: 143.3}, {X: 98.0, Y: 143.3}, {X: 99.0, Y: 144.3}, {X: 99.0, Y: 145.3}, {X: 100.0, Y: 146.3}, {X: 100.0, Y: 148.3}, {X: 101.0, Y: 150.3}, {X: 101.0, Y: 152.3}, {X: 100.0, Y: 154.3}, {X: 100.0, Y: 156.3}, {X: 99.0, Y: 158.3}, {X: 98.0, Y: 160.3}, {X: 97.0, Y: 162.3}, {X: 95.0, Y: 164.3}, {X: 94.0, Y: 165.3}, {X: 93.0, Y: 166.3}, {X: 92.0, Y: 167.3}, {X: 91.0, Y: 168.3}, {X: 90.0, Y: 168.3}, {X: 89.0, Y: 168.3}, {X: 88.0, Y: 168.3}, {X: 88.0, Y: 167.3}, {X: 89.0, Y: 167.3}, {X: 90.0, Y: 167.3}, {X: 91.0, Y: 168.3}, {X: 93.0, Y: 168.3}, {X: 94.0, Y: 169.3}, {X: 96.0, Y: 170.3}, {X: 97.0, Y: 171.3}, {X: 99.0, Y: 173.3}, {X: 100.0, Y: 175.3}, {X: 101.0, Y: 177.3}, {X: 103.0, Y: 179.3}, {X: 104.0, Y: 180.3}, {X: 105.0, Y: 182.3}, {X: 105.0, Y: 184.3}, {X: 106.0, Y: 185.3}, {X: 107.0, Y: 186.3}, {X: 107.0, Y: 187.3}, {X: 107.0, Y: 188.3}, {X: 107.0, Y: 189.3}, {X: 108.0, Y: 189.3}, {X: 107.0, Y: 189.3}, {X: 107.0, Y: 188.3}, {X: 107.0, Y: 187.3}, {X: 107.0, Y: 186.3}}

	fitteds := fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 2)
}

func TestF(t *testing.T) {
	points := Shape{{X: 100.0, Y: 139.3}, {X: 100.0, Y: 138.3}, {X: 100.0, Y: 137.3}, {X: 100.0, Y: 136.3}, {X: 100.0, Y: 135.3}, {X: 100.0, Y: 134.3}, {X: 100.0, Y: 133.3}, {X: 100.0, Y: 132.3}, {X: 100.0, Y: 133.3}, {X: 100.0, Y: 134.3}, {X: 100.0, Y: 135.3}, {X: 101.0, Y: 138.3}, {X: 101.0, Y: 141.3}, {X: 101.0, Y: 146.3}, {X: 101.0, Y: 152.3}, {X: 102.0, Y: 158.3}, {X: 102.0, Y: 163.3}, {X: 102.0, Y: 166.3}, {X: 102.0, Y: 169.3}, {X: 102.0, Y: 171.3}, {X: 102.0, Y: 173.3}, {X: 103.0, Y: 176.3}, {X: 103.0, Y: 179.3}, {X: 103.0, Y: 182.3}, {X: 103.0, Y: 184.3}, {X: 103.0, Y: 185.3}, {X: 103.0, Y: 186.3}, {X: 103.0, Y: 187.3}, {X: 102.0, Y: 186.3}}

	fitteds := fitCubicBeziers(points)
	tu.AssertEqual(t, len(mergeSimilarCurves(fitteds)), 1)
}

func Test1(t *testing.T) {
	points := Shape{
		{X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 45.0, Y: 58.0}, {X: 46.0, Y: 58.0}, {X: 47.0, Y: 58.0}, {X: 48.0, Y: 57.0}, {X: 50.0, Y: 56.0}, {X: 52.0, Y: 55.0}, {X: 53.0, Y: 54.0}, {X: 55.0, Y: 53.0}, {X: 56.0, Y: 51.0}, {X: 57.0, Y: 49.0}, {X: 59.0, Y: 48.0}, {X: 59.0, Y: 46.0}, {X: 60.0, Y: 44.0}, {X: 60.0, Y: 43.0}, {X: 60.0, Y: 42.0}, {X: 61.0, Y: 41.0}, {X: 61.0, Y: 40.0}, {X: 61.0, Y: 40.0}, {X: 61.0, Y: 40.0}, {X: 61.0, Y: 40.0}, {X: 61.0, Y: 40.0}, {X: 61.0, Y: 41.0}, {X: 60.0, Y: 43.0}, {X: 60.0, Y: 45.0}, {X: 60.0, Y: 48.0}, {X: 59.0, Y: 52.0}, {X: 59.0, Y: 55.0}, {X: 58.0, Y: 59.0}, {X: 58.0, Y: 63.0}, {X: 57.0, Y: 66.0}, {X: 57.0, Y: 69.0}, {X: 57.0, Y: 71.0}, {X: 57.0, Y: 73.0}, {X: 57.0, Y: 75.0}, {X: 57.0, Y: 75.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 76.0}, {X: 57.0, Y: 75.0},
	}
	fitteds := fitCubicBeziers(points)
	fitteds = mergeSimilarCurves(fitteds)
	tu.AssertEqual(t, len(fitteds), 2)

	printShape(t, fitteds[0].toPoints(), "c0")
	printShape(t, fitteds[1].toPoints(), "c1")
}

func Test_areBeziersSpuriousCurvature(t *testing.T) {
	c1 := Bezier{Pos{X: 45.0, Y: 58.0}, Pos{X: 59.2, Y: 58.0}, Pos{X: 61.5, Y: 29.9}, Pos{X: 60.0, Y: 48.0}}
	c2 := Bezier{Pos{X: 60.0, Y: 48.0}, Pos{X: 58.5, Y: 62.0}, Pos{X: 58.5, Y: 62.0}, Pos{X: 57.0, Y: 76.0}}
	f1, f2, ok := areBeziersSpuriousCurvature(c1, c2)
	tu.Assert(t, ok)

	printShape(t, f1.toPoints(), "1")
	printShape(t, f2.toPoints(), "2")
}

func TestRemoveRepetitions(t *testing.T) {
	A := Symbol{
		{{X: 234.0, Y: 22.0}, {X: 234.0, Y: 22.0}, {X: 234.0, Y: 22.0}, {X: 234.0, Y: 22.0}, {X: 234.0, Y: 22.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 234.0, Y: 21.0}, {X: 233.0, Y: 23.0}, {X: 233.0, Y: 24.0}, {X: 232.0, Y: 27.0}, {X: 231.0, Y: 30.0}, {X: 229.0, Y: 34.0}, {X: 228.0, Y: 37.0}, {X: 227.0, Y: 41.0}, {X: 226.0, Y: 45.0}, {X: 225.0, Y: 48.0}, {X: 224.0, Y: 52.0}, {X: 223.0, Y: 55.0}, {X: 222.0, Y: 57.0}, {X: 221.0, Y: 59.0}, {X: 221.0, Y: 61.0}, {X: 221.0, Y: 61.0}, {X: 221.0, Y: 62.0}, {X: 221.0, Y: 62.0}, {X: 221.0, Y: 62.0}, {X: 221.0, Y: 62.0}, {X: 221.0, Y: 61.0}, {X: 221.0, Y: 61.0}, {X: 222.0, Y: 59.0}, {X: 223.0, Y: 57.0}, {X: 224.0, Y: 54.0}, {X: 225.0, Y: 51.0}, {X: 226.0, Y: 48.0}, {X: 227.0, Y: 45.0}, {X: 228.0, Y: 42.0}, {X: 228.0, Y: 40.0}, {X: 229.0, Y: 37.0}, {X: 230.0, Y: 34.0}, {X: 230.0, Y: 31.0}, {X: 231.0, Y: 28.0}, {X: 232.0, Y: 25.0}, {X: 232.0, Y: 22.0}, {X: 232.0, Y: 20.0}, {X: 233.0, Y: 18.0}, {X: 233.0, Y: 17.0}, {X: 233.0, Y: 16.0}, {X: 233.0, Y: 16.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 233.0, Y: 15.0}, {X: 234.0, Y: 16.0}, {X: 235.0, Y: 17.0}, {X: 236.0, Y: 18.0}, {X: 237.0, Y: 20.0}, {X: 238.0, Y: 22.0}, {X: 240.0, Y: 25.0}, {X: 241.0, Y: 27.0}, {X: 242.0, Y: 30.0}, {X: 243.0, Y: 33.0}, {X: 244.0, Y: 36.0}, {X: 245.0, Y: 39.0}, {X: 246.0, Y: 42.0}, {X: 247.0, Y: 44.0}, {X: 248.0, Y: 47.0}, {X: 249.0, Y: 49.0}, {X: 249.0, Y: 51.0}, {X: 250.0, Y: 53.0}, {X: 250.0, Y: 55.0}, {X: 251.0, Y: 56.0}, {X: 251.0, Y: 57.0}, {X: 252.0, Y: 59.0}, {X: 252.0, Y: 59.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}, {X: 252.0, Y: 60.0}},
	}.Footprint().Strokes[0].Curves
	tu.AssertEqual(t, len(A), 2) // and not 3
}
