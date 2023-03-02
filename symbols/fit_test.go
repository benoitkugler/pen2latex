package symbols

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
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

func minMax2D(values []Pos) image.Rectangle {
	minX, maxX := values[0].X, values[0].X
	minY, maxY := values[0].Y, values[0].Y
	for _, v := range values {
		if v.X > maxX {
			maxX = v.X
		}
		if v.X < minX {
			minX = v.X
		}
		if v.Y > maxY {
			maxY = v.Y
		}
		if v.Y < minY {
			minY = v.Y
		}
	}
	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

// return a rough version of a graph (sh[i].X, sh[i].Y)
func graph2D(sh Shape) image.Image {
	// adjust the graph height
	rect := minMax2D(sh)
	out := image.NewGray(rect)
	for _, v := range sh {
		out.SetGray(int(v.X), int(v.Y), color.Gray{255})
	}
	return out
}

func generateBezierPoints(b Bezier) Shape {
	var sh Shape
	for t := range [20]int{} {
		sh = append(sh, b.eval(fl(t)/20))
	}
	return sh
}

func printShape(t *testing.T, sh Shape, filename string) Shape {
	graph := graph2D(sh)
	f, err := os.Create(filename)
	tu.AssertNoErr(t, err)

	err = png.Encode(f, graph)
	tu.AssertNoErr(t, err)

	fmt.Printf("Graph saved in file://%s\n", filename)

	return sh
}

func circle(center Pos, radius fl) Shape {
	var out Shape
	for theta := range [30]int{} {
		out = append(out, center.Add(Pos{
			X: float32(math.Cos(2 * math.Pi * float64(theta) / 30)),
			Y: float32(math.Sin(2 * math.Pi * float64(theta) / 30)),
		}.ScaleTo(radius)))
	}
	return out
}

func ellipse(center Pos, ra, rb fl) Shape {
	s := circle(Pos{}, 1)
	for i, p := range s {
		s[i] = Pos{ra * p.X, rb * p.Y}.Add(center)
	}
	return s
}

func TestFitBezier(t *testing.T) {
	tmpDir := os.TempDir()

	origin := Bezier{Pos{}, Pos{30, 40}, Pos{50, -40}, Pos{60, 0}}
	points := generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube1_origin.png"))

	fitted, _ := fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube1_fitted.png"))

	origin = Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	points = generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube2_origin.png"))

	fitted, _ = fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube2_fitted.png"))

	// linear shape
	origin = Bezier{Pos{}, Pos{30, 30}, Pos{40, 40}, Pos{60, 60}}
	points = generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube3_origin.png"))

	fitted, _ = fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube3_fitted.png"))

	_, errLine := fitSegment(points)
	tu.Assert(t, errLine == 0)

	// data from a circle, not a bezier !
	points = circle(Pos{30, 30}, 20)
	fitted, errCube := fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube4_fitted.png"))
	_, errCircle := fitCircle(points)
	tu.Assert(t, errCircle <= errCube)
}

func almostEqual(u, v fl) bool     { return abs(u-v) < 1e-4 }
func almostEqualPos(u, v Pos) bool { return almostEqual(u.X, v.X) && almostEqual(u.Y, v.Y) }

func TestFitEllipse(t *testing.T) {
	got, dist := fitEllipse(circle(Pos{30, 30}, 20))
	tu.Assert(t, almostEqual(dist, 0))
	tu.Assert(t, almostEqualPos(got.Center, Pos{30, 30}) && almostEqualPos(got.Radius, Pos{20, 20}))

	got, dist = fitEllipse(ellipse(Pos{5, 5}, 20, 10))
	tu.Assert(t, almostEqual(dist, 0))
	tu.Assert(t, almostEqualPos(got.Center, Pos{5, 5}) && almostEqualPos(got.Radius, Pos{20, 10}))
}

func TestIdentify(t *testing.T) {
	points := circle(Pos{30, 30}, 20)
	tu.Assert(t, points.identify().Kind() == SAKCircle)

	origin := Bezier{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	points = generateBezierPoints(origin)
	tu.Assert(t, points.identify().Kind() == SAKBezier)
}

var realShapes = []struct {
	expected []ShapeAtomKind
	inputs   []Shape
}{
	// line shape
	{
		[]ShapeAtomKind{SAKSegment},
		[]Shape{
			{
				{X: 58.0, Y: 39.3}, {X: 59.0, Y: 38.3}, {X: 60.0, Y: 37.3}, {X: 61.0, Y: 36.3}, {X: 61.0, Y: 35.3}, {X: 62.0, Y: 34.3}, {X: 63.0, Y: 33.3}, {X: 64.0, Y: 32.3}, {X: 65.0, Y: 31.3}, {X: 65.0, Y: 30.3}, {X: 66.0, Y: 29.3}, {X: 67.0, Y: 28.3}, {X: 68.0, Y: 27.3}, {X: 68.0, Y: 26.3}, {X: 69.0, Y: 25.3}, {X: 70.0, Y: 24.3}, {X: 71.0, Y: 23.3}, {X: 72.0, Y: 22.3}, {X: 72.0, Y: 21.3}, {X: 73.0, Y: 20.3}, {X: 74.0, Y: 19.3}, {X: 75.0, Y: 18.3}, {X: 75.0, Y: 17.3}, {X: 76.0, Y: 17.3}, {X: 76.0, Y: 16.3}, {X: 77.0, Y: 16.3}, {X: 77.0, Y: 15.3}, {X: 78.0, Y: 14.3}, {X: 79.0, Y: 13.3}, {X: 80.0, Y: 12.3},
			},
		},
	},

	// one bezier curve
	{
		[]ShapeAtomKind{SAKBezier},
		[]Shape{
			// openParShape
			{
				{X: 72.0, Y: 9.3}, {X: 72.0, Y: 10.3}, {X: 72.0, Y: 11.3}, {X: 71.0, Y: 12.3}, {X: 71.0, Y: 13.3}, {X: 71.0, Y: 14.3}, {X: 70.0, Y: 15.3}, {X: 70.0, Y: 16.3}, {X: 70.0, Y: 17.3}, {X: 70.0, Y: 18.3}, {X: 69.0, Y: 19.3}, {X: 69.0, Y: 20.3}, {X: 69.0, Y: 21.3}, {X: 69.0, Y: 22.3}, {X: 69.0, Y: 23.3}, {X: 69.0, Y: 24.3}, {X: 69.0, Y: 25.3}, {X: 70.0, Y: 26.3}, {X: 70.0, Y: 27.3}, {X: 70.0, Y: 28.3}, {X: 70.0, Y: 29.3}, {X: 71.0, Y: 30.3}, {X: 71.0, Y: 31.3}, {X: 72.0, Y: 32.3}, {X: 72.0, Y: 33.3}, {X: 73.0, Y: 34.3}, {X: 73.0, Y: 35.3}, {X: 74.0, Y: 36.3}, {X: 74.0, Y: 37.3}, {X: 75.0, Y: 37.3}, {X: 75.0, Y: 38.3}, {X: 76.0, Y: 38.3}, {X: 77.0, Y: 39.3}, {X: 77.0, Y: 40.3}, {X: 78.0, Y: 40.3}, {X: 78.0, Y: 41.3}, {X: 79.0, Y: 41.3}, {X: 80.0, Y: 41.3}, {X: 80.0, Y: 42.3}, {X: 81.0, Y: 42.3}, {X: 81.0, Y: 41.3}, {X: 81.0, Y: 40.3},
			},

			// CShape
			{
				{X: 149.0, Y: 8.3}, {X: 148.0, Y: 8.3}, {X: 148.0, Y: 9.3}, {X: 147.0, Y: 9.3}, {X: 147.0, Y: 10.3}, {X: 146.0, Y: 10.3}, {X: 145.0, Y: 11.3}, {X: 145.0, Y: 12.3}, {X: 144.0, Y: 13.3}, {X: 143.0, Y: 14.3}, {X: 143.0, Y: 15.3}, {X: 142.0, Y: 16.3}, {X: 142.0, Y: 17.3}, {X: 141.0, Y: 18.3}, {X: 141.0, Y: 19.3}, {X: 141.0, Y: 20.3}, {X: 140.0, Y: 21.3}, {X: 140.0, Y: 22.3}, {X: 140.0, Y: 23.3}, {X: 140.0, Y: 24.3}, {X: 140.0, Y: 25.3}, {X: 140.0, Y: 26.3}, {X: 140.0, Y: 27.3}, {X: 140.0, Y: 28.3}, {X: 141.0, Y: 29.3}, {X: 141.0, Y: 30.3}, {X: 142.0, Y: 31.3}, {X: 143.0, Y: 32.3}, {X: 144.0, Y: 32.3}, {X: 145.0, Y: 33.3}, {X: 146.0, Y: 33.3}, {X: 147.0, Y: 33.3}, {X: 148.0, Y: 34.3}, {X: 149.0, Y: 34.3}, {X: 150.0, Y: 34.3}, {X: 151.0, Y: 35.3}, {X: 153.0, Y: 35.3}, {X: 154.0, Y: 35.3}, {X: 155.0, Y: 35.3}, {X: 156.0, Y: 35.3}, {X: 157.0, Y: 35.3}, {X: 158.0, Y: 35.3}, {X: 159.0, Y: 35.3}, {X: 159.0, Y: 36.3},
			},
			// )
			{
				{X: 98.0, Y: 9.3}, {X: 99.0, Y: 10.3}, {X: 100.0, Y: 11.3}, {X: 101.0, Y: 12.3}, {X: 102.0, Y: 13.3}, {X: 102.0, Y: 14.3}, {X: 103.0, Y: 15.3}, {X: 103.0, Y: 16.3}, {X: 104.0, Y: 17.3}, {X: 104.0, Y: 18.3}, {X: 104.0, Y: 19.3}, {X: 104.0, Y: 20.3}, {X: 105.0, Y: 21.3}, {X: 105.0, Y: 22.3}, {X: 105.0, Y: 23.3}, {X: 105.0, Y: 24.3}, {X: 105.0, Y: 25.3}, {X: 105.0, Y: 27.3}, {X: 105.0, Y: 29.3}, {X: 105.0, Y: 30.3}, {X: 105.0, Y: 32.3}, {X: 105.0, Y: 34.3}, {X: 105.0, Y: 35.3}, {X: 105.0, Y: 36.3}, {X: 105.0, Y: 38.3}, {X: 104.0, Y: 38.3}, {X: 104.0, Y: 39.3},
			},
			// t (without bar)
			{
				{X: 87.0, Y: 13.3}, {X: 87.0, Y: 14.3}, {X: 87.0, Y: 15.3}, {X: 87.0, Y: 16.3}, {X: 87.0, Y: 18.3}, {X: 87.0, Y: 19.3}, {X: 86.0, Y: 20.3}, {X: 86.0, Y: 21.3}, {X: 86.0, Y: 23.3}, {X: 86.0, Y: 24.3}, {X: 86.0, Y: 25.3}, {X: 86.0, Y: 27.3}, {X: 86.0, Y: 28.3}, {X: 86.0, Y: 30.3}, {X: 86.0, Y: 31.3}, {X: 86.0, Y: 33.3}, {X: 86.0, Y: 34.3}, {X: 86.0, Y: 35.3}, {X: 86.0, Y: 36.3}, {X: 86.0, Y: 37.3}, {X: 86.0, Y: 38.3}, {X: 86.0, Y: 39.3}, {X: 87.0, Y: 40.3}, {X: 88.0, Y: 42.3}, {X: 89.0, Y: 42.3}, {X: 89.0, Y: 43.3}, {X: 90.0, Y: 43.3}, {X: 91.0, Y: 43.3}, {X: 92.0, Y: 43.3}, {X: 92.0, Y: 42.3}, {X: 93.0, Y: 42.3}, {X: 93.0, Y: 41.3}, {X: 94.0, Y: 41.3},
			},
			// integral
			{
				{X: 115.0, Y: 7.3}, {X: 114.0, Y: 8.3}, {X: 114.0, Y: 9.3}, {X: 113.0, Y: 10.3}, {X: 113.0, Y: 11.3}, {X: 113.0, Y: 12.3}, {X: 113.0, Y: 13.3}, {X: 113.0, Y: 14.3}, {X: 113.0, Y: 15.3}, {X: 113.0, Y: 16.3}, {X: 113.0, Y: 18.3}, {X: 113.0, Y: 19.3}, {X: 113.0, Y: 20.3}, {X: 113.0, Y: 21.3}, {X: 113.0, Y: 22.3}, {X: 113.0, Y: 23.3}, {X: 114.0, Y: 24.3}, {X: 114.0, Y: 25.3}, {X: 114.0, Y: 26.3}, {X: 114.0, Y: 27.3}, {X: 114.0, Y: 28.3}, {X: 114.0, Y: 29.3}, {X: 114.0, Y: 31.3}, {X: 114.0, Y: 32.3}, {X: 113.0, Y: 33.3}, {X: 113.0, Y: 35.3}, {X: 112.0, Y: 37.3}, {X: 111.0, Y: 39.3}, {X: 110.0, Y: 40.3}, {X: 109.0, Y: 42.3}, {X: 108.0, Y: 43.3}, {X: 107.0, Y: 44.3}, {X: 105.0, Y: 46.3}, {X: 104.0, Y: 46.3}, {X: 104.0, Y: 47.3}, {X: 103.0, Y: 48.3}, {X: 102.0, Y: 48.3},
			},
			// S
			{
				{X: 113.0, Y: 4.3}, {X: 112.0, Y: 4.3}, {X: 111.0, Y: 4.3}, {X: 110.0, Y: 4.3}, {X: 109.0, Y: 5.3}, {X: 108.0, Y: 5.3}, {X: 107.0, Y: 5.3}, {X: 106.0, Y: 6.3}, {X: 105.0, Y: 7.3}, {X: 104.0, Y: 8.3}, {X: 103.0, Y: 9.3}, {X: 103.0, Y: 10.3}, {X: 102.0, Y: 11.3}, {X: 102.0, Y: 12.3}, {X: 102.0, Y: 13.3}, {X: 101.0, Y: 14.3}, {X: 101.0, Y: 15.3}, {X: 101.0, Y: 16.3}, {X: 102.0, Y: 16.3}, {X: 103.0, Y: 17.3}, {X: 104.0, Y: 17.3}, {X: 105.0, Y: 17.3}, {X: 106.0, Y: 17.3}, {X: 107.0, Y: 17.3}, {X: 108.0, Y: 17.3}, {X: 109.0, Y: 17.3}, {X: 110.0, Y: 17.3}, {X: 111.0, Y: 17.3}, {X: 112.0, Y: 17.3}, {X: 113.0, Y: 17.3}, {X: 114.0, Y: 17.3}, {X: 115.0, Y: 17.3}, {X: 116.0, Y: 18.3}, {X: 117.0, Y: 18.3}, {X: 117.0, Y: 19.3}, {X: 117.0, Y: 20.3}, {X: 117.0, Y: 21.3}, {X: 117.0, Y: 22.3}, {X: 117.0, Y: 23.3}, {X: 116.0, Y: 24.3}, {X: 115.0, Y: 24.3}, {X: 115.0, Y: 25.3}, {X: 114.0, Y: 25.3}, {X: 113.0, Y: 26.3}, {X: 112.0, Y: 26.3}, {X: 111.0, Y: 26.3}, {X: 110.0, Y: 27.3}, {X: 108.0, Y: 27.3}, {X: 107.0, Y: 27.3}, {X: 106.0, Y: 27.3}, {X: 105.0, Y: 27.3}, {X: 104.0, Y: 27.3}, {X: 103.0, Y: 27.3},
			},
		},
	},

	// two lines
	{
		[]ShapeAtomKind{SAKSegment, SAKSegment},
		[]Shape{
			// LShape
			{
				{X: 74.0, Y: 13.3}, {X: 74.0, Y: 14.3}, {X: 74.0, Y: 15.3}, {X: 74.0, Y: 16.3}, {X: 74.0, Y: 17.3}, {X: 75.0, Y: 18.3}, {X: 75.0, Y: 19.3}, {X: 75.0, Y: 20.3}, {X: 75.0, Y: 21.3}, {X: 75.0, Y: 22.3}, {X: 75.0, Y: 24.3}, {X: 75.0, Y: 25.3}, {X: 75.0, Y: 26.3}, {X: 75.0, Y: 27.3}, {X: 75.0, Y: 28.3}, {X: 75.0, Y: 29.3}, {X: 75.0, Y: 30.3}, {X: 75.0, Y: 31.3}, {X: 75.0, Y: 32.3}, {X: 75.0, Y: 33.3}, {X: 75.0, Y: 34.3}, {X: 75.0, Y: 35.3}, {X: 75.0, Y: 36.3}, {X: 75.0, Y: 37.3}, {X: 75.0, Y: 38.3}, {X: 75.0, Y: 39.3}, {X: 76.0, Y: 39.3}, {X: 77.0, Y: 39.3}, {X: 78.0, Y: 39.3}, {X: 79.0, Y: 39.3}, {X: 80.0, Y: 39.3}, {X: 81.0, Y: 39.3}, {X: 82.0, Y: 39.3}, {X: 83.0, Y: 39.3}, {X: 84.0, Y: 39.3}, {X: 85.0, Y: 39.3}, {X: 86.0, Y: 39.3}, {X: 87.0, Y: 39.3}, {X: 88.0, Y: 39.3}, {X: 89.0, Y: 39.3}, {X: 90.0, Y: 39.3}, {X: 91.0, Y: 39.3}, {X: 92.0, Y: 39.3}, {X: 92.0, Y: 38.3}, {X: 93.0, Y: 38.3}, {X: 94.0, Y: 38.3}, {X: 93.0, Y: 40.3},
			},
			// VShape
			{
				{X: 121.0, Y: 15.3}, {X: 122.0, Y: 16.3}, {X: 123.0, Y: 17.3}, {X: 124.0, Y: 19.3}, {X: 126.0, Y: 20.3}, {X: 127.0, Y: 22.3}, {X: 128.0, Y: 23.3}, {X: 129.0, Y: 25.3}, {X: 130.0, Y: 27.3}, {X: 132.0, Y: 29.3}, {X: 133.0, Y: 31.3}, {X: 134.0, Y: 33.3}, {X: 135.0, Y: 34.3}, {X: 136.0, Y: 35.3}, {X: 137.0, Y: 37.3}, {X: 138.0, Y: 38.3}, {X: 138.0, Y: 39.3}, {X: 139.0, Y: 39.3}, {X: 139.0, Y: 38.3}, {X: 139.0, Y: 37.3}, {X: 140.0, Y: 36.3}, {X: 141.0, Y: 35.3}, {X: 142.0, Y: 33.3}, {X: 143.0, Y: 31.3}, {X: 144.0, Y: 29.3}, {X: 145.0, Y: 27.3}, {X: 146.0, Y: 25.3}, {X: 147.0, Y: 23.3}, {X: 148.0, Y: 22.3}, {X: 149.0, Y: 20.3}, {X: 150.0, Y: 18.3}, {X: 151.0, Y: 17.3}, {X: 152.0, Y: 15.3}, {X: 153.0, Y: 14.3}, {X: 153.0, Y: 13.3}, {X: 154.0, Y: 11.3}, {X: 155.0, Y: 11.3}, {X: 155.0, Y: 10.3}, {X: 155.0, Y: 9.3}, {X: 156.0, Y: 9.3}, {X: 156.0, Y: 8.3}, {X: 157.0, Y: 8.3}, {X: 156.0, Y: 8.3},
			},
			// A without horizontal
			{
				{X: 63.0, Y: 36.3}, {X: 63.0, Y: 35.3}, {X: 64.0, Y: 35.3}, {X: 64.0, Y: 34.3}, {X: 65.0, Y: 33.3}, {X: 65.0, Y: 32.3}, {X: 66.0, Y: 31.3}, {X: 67.0, Y: 30.3}, {X: 67.0, Y: 29.3}, {X: 68.0, Y: 29.3}, {X: 68.0, Y: 28.3}, {X: 69.0, Y: 27.3}, {X: 69.0, Y: 26.3}, {X: 70.0, Y: 25.3}, {X: 71.0, Y: 24.3}, {X: 71.0, Y: 23.3}, {X: 72.0, Y: 22.3}, {X: 72.0, Y: 21.3}, {X: 73.0, Y: 20.3}, {X: 73.0, Y: 19.3}, {X: 74.0, Y: 19.3}, {X: 74.0, Y: 18.3}, {X: 75.0, Y: 17.3}, {X: 75.0, Y: 16.3}, {X: 75.0, Y: 15.3}, {X: 76.0, Y: 14.3}, {X: 76.0, Y: 13.3}, {X: 77.0, Y: 12.3}, {X: 77.0, Y: 11.3}, {X: 78.0, Y: 11.3}, {X: 78.0, Y: 10.3}, {X: 79.0, Y: 9.3}, {X: 79.0, Y: 8.3}, {X: 80.0, Y: 8.3}, {X: 80.0, Y: 7.3}, {X: 80.0, Y: 8.3}, {X: 81.0, Y: 8.3}, {X: 81.0, Y: 9.3}, {X: 82.0, Y: 10.3}, {X: 83.0, Y: 11.3}, {X: 84.0, Y: 12.3}, {X: 85.0, Y: 13.3}, {X: 85.0, Y: 14.3}, {X: 86.0, Y: 15.3}, {X: 87.0, Y: 16.3}, {X: 88.0, Y: 18.3}, {X: 89.0, Y: 19.3}, {X: 90.0, Y: 20.3}, {X: 91.0, Y: 21.3}, {X: 91.0, Y: 22.3}, {X: 92.0, Y: 23.3}, {X: 93.0, Y: 24.3}, {X: 94.0, Y: 24.3}, {X: 95.0, Y: 25.3}, {X: 96.0, Y: 26.3}, {X: 96.0, Y: 27.3}, {X: 97.0, Y: 28.3}, {X: 98.0, Y: 29.3}, {X: 98.0, Y: 30.3}, {X: 99.0, Y: 31.3}, {X: 100.0, Y: 32.3}, {X: 101.0, Y: 32.3}, {X: 102.0, Y: 33.3}, {X: 103.0, Y: 34.3}, {X: 103.0, Y: 35.3}, {X: 104.0, Y: 36.3}, {X: 103.0, Y: 37.3},
			},
			// same with more noise
			{
				{X: 70.0, Y: 45.3}, {X: 70.0, Y: 44.3}, {X: 71.0, Y: 44.3}, {X: 71.0, Y: 43.3}, {X: 71.0, Y: 42.3}, {X: 72.0, Y: 41.3}, {X: 73.0, Y: 39.3}, {X: 73.0, Y: 38.3}, {X: 74.0, Y: 36.3}, {X: 75.0, Y: 34.3}, {X: 76.0, Y: 32.3}, {X: 77.0, Y: 30.3}, {X: 78.0, Y: 28.3}, {X: 79.0, Y: 27.3}, {X: 80.0, Y: 25.3}, {X: 81.0, Y: 23.3}, {X: 82.0, Y: 22.3}, {X: 82.0, Y: 21.3}, {X: 82.0, Y: 20.3}, {X: 83.0, Y: 19.3}, {X: 83.0, Y: 18.3}, {X: 84.0, Y: 17.3}, {X: 84.0, Y: 15.3}, {X: 85.0, Y: 15.3}, {X: 85.0, Y: 14.3}, {X: 85.0, Y: 13.3}, {X: 86.0, Y: 14.3}, {X: 86.0, Y: 15.3}, {X: 87.0, Y: 15.3}, {X: 88.0, Y: 17.3}, {X: 88.0, Y: 18.3}, {X: 89.0, Y: 19.3}, {X: 91.0, Y: 21.3}, {X: 92.0, Y: 22.3}, {X: 93.0, Y: 24.3}, {X: 94.0, Y: 26.3}, {X: 95.0, Y: 28.3}, {X: 96.0, Y: 29.3}, {X: 97.0, Y: 31.3}, {X: 98.0, Y: 33.3}, {X: 99.0, Y: 34.3}, {X: 100.0, Y: 35.3}, {X: 101.0, Y: 37.3}, {X: 102.0, Y: 38.3}, {X: 102.0, Y: 39.3}, {X: 103.0, Y: 40.3}, {X: 104.0, Y: 41.3}, {X: 104.0, Y: 40.3}, {X: 103.0, Y: 39.3},
			},
			{
				{X: 77.0, Y: 43.3}, {X: 78.0, Y: 42.3}, {X: 79.0, Y: 41.3}, {X: 79.0, Y: 39.3}, {X: 80.0, Y: 38.3}, {X: 81.0, Y: 36.3}, {X: 82.0, Y: 34.3}, {X: 83.0, Y: 31.3}, {X: 84.0, Y: 29.3}, {X: 85.0, Y: 27.3}, {X: 86.0, Y: 24.3}, {X: 87.0, Y: 22.3}, {X: 88.0, Y: 21.3}, {X: 88.0, Y: 19.3}, {X: 88.0, Y: 18.3}, {X: 89.0, Y: 17.3}, {X: 89.0, Y: 16.3}, {X: 89.0, Y: 15.3}, {X: 89.0, Y: 14.3}, {X: 89.0, Y: 13.3}, {X: 89.0, Y: 12.3}, {X: 89.0, Y: 11.3}, {X: 90.0, Y: 11.3}, {X: 91.0, Y: 12.3}, {X: 91.0, Y: 13.3}, {X: 92.0, Y: 15.3}, {X: 93.0, Y: 16.3}, {X: 94.0, Y: 18.3}, {X: 95.0, Y: 20.3}, {X: 96.0, Y: 22.3}, {X: 97.0, Y: 24.3}, {X: 98.0, Y: 26.3}, {X: 100.0, Y: 28.3}, {X: 101.0, Y: 30.3}, {X: 102.0, Y: 32.3}, {X: 103.0, Y: 34.3}, {X: 103.0, Y: 35.3}, {X: 104.0, Y: 35.3}, {X: 104.0, Y: 36.3}, {X: 104.0, Y: 35.3},
			},
		},
	},

	// one circle
	{
		[]ShapeAtomKind{SAKCircle},
		[]Shape{
			// o
			{
				{X: 114.0, Y: 16.3}, {X: 114.0, Y: 17.3}, {X: 114.0, Y: 18.3}, {X: 113.0, Y: 19.3}, {X: 113.0, Y: 20.3}, {X: 113.0, Y: 21.3}, {X: 113.0, Y: 22.3}, {X: 113.0, Y: 23.3}, {X: 113.0, Y: 24.3}, {X: 114.0, Y: 25.3}, {X: 115.0, Y: 25.3}, {X: 115.0, Y: 26.3}, {X: 116.0, Y: 26.3}, {X: 117.0, Y: 26.3}, {X: 118.0, Y: 27.3}, {X: 119.0, Y: 27.3}, {X: 120.0, Y: 27.3}, {X: 121.0, Y: 27.3}, {X: 122.0, Y: 27.3}, {X: 123.0, Y: 27.3}, {X: 124.0, Y: 27.3}, {X: 125.0, Y: 27.3}, {X: 126.0, Y: 27.3}, {X: 127.0, Y: 27.3}, {X: 127.0, Y: 26.3}, {X: 128.0, Y: 26.3}, {X: 128.0, Y: 25.3}, {X: 129.0, Y: 25.3}, {X: 129.0, Y: 24.3}, {X: 129.0, Y: 23.3}, {X: 130.0, Y: 22.3}, {X: 130.0, Y: 21.3}, {X: 130.0, Y: 20.3}, {X: 130.0, Y: 19.3}, {X: 129.0, Y: 19.3}, {X: 129.0, Y: 18.3}, {X: 128.0, Y: 17.3}, {X: 127.0, Y: 17.3}, {X: 125.0, Y: 16.3}, {X: 124.0, Y: 16.3}, {X: 123.0, Y: 16.3}, {X: 121.0, Y: 16.3}, {X: 120.0, Y: 16.3}, {X: 118.0, Y: 16.3}, {X: 117.0, Y: 17.3}, {X: 116.0, Y: 17.3}, {X: 115.0, Y: 18.3}, {X: 115.0, Y: 19.3}, {X: 114.0, Y: 20.3},
			},

			// O
			{
				{X: 90.0, Y: 8.3}, {X: 89.0, Y: 8.3}, {X: 88.0, Y: 8.3}, {X: 87.0, Y: 9.3}, {X: 86.0, Y: 9.3}, {X: 85.0, Y: 9.3}, {X: 84.0, Y: 10.3}, {X: 83.0, Y: 10.3}, {X: 82.0, Y: 11.3}, {X: 81.0, Y: 11.3}, {X: 80.0, Y: 12.3}, {X: 79.0, Y: 12.3}, {X: 79.0, Y: 13.3}, {X: 78.0, Y: 13.3}, {X: 77.0, Y: 14.3}, {X: 77.0, Y: 15.3}, {X: 76.0, Y: 16.3}, {X: 76.0, Y: 17.3}, {X: 76.0, Y: 18.3}, {X: 76.0, Y: 19.3}, {X: 75.0, Y: 19.3}, {X: 75.0, Y: 20.3}, {X: 75.0, Y: 21.3}, {X: 76.0, Y: 22.3}, {X: 76.0, Y: 23.3}, {X: 76.0, Y: 24.3}, {X: 77.0, Y: 25.3}, {X: 78.0, Y: 26.3}, {X: 78.0, Y: 27.3}, {X: 78.0, Y: 28.3}, {X: 79.0, Y: 29.3}, {X: 80.0, Y: 29.3}, {X: 80.0, Y: 30.3}, {X: 81.0, Y: 31.3}, {X: 82.0, Y: 32.3}, {X: 83.0, Y: 33.3}, {X: 84.0, Y: 33.3}, {X: 84.0, Y: 34.3}, {X: 85.0, Y: 34.3}, {X: 86.0, Y: 34.3}, {X: 87.0, Y: 35.3}, {X: 88.0, Y: 35.3}, {X: 89.0, Y: 35.3}, {X: 90.0, Y: 35.3}, {X: 91.0, Y: 36.3}, {X: 92.0, Y: 36.3}, {X: 93.0, Y: 36.3}, {X: 94.0, Y: 35.3}, {X: 95.0, Y: 35.3}, {X: 96.0, Y: 34.3}, {X: 97.0, Y: 34.3}, {X: 98.0, Y: 33.3}, {X: 100.0, Y: 32.3}, {X: 100.0, Y: 31.3}, {X: 101.0, Y: 30.3}, {X: 102.0, Y: 29.3}, {X: 102.0, Y: 28.3}, {X: 103.0, Y: 27.3}, {X: 104.0, Y: 26.3}, {X: 104.0, Y: 25.3}, {X: 104.0, Y: 24.3}, {X: 104.0, Y: 22.3}, {X: 105.0, Y: 21.3}, {X: 105.0, Y: 20.3}, {X: 104.0, Y: 19.3}, {X: 104.0, Y: 18.3}, {X: 103.0, Y: 16.3}, {X: 102.0, Y: 15.3}, {X: 100.0, Y: 14.3}, {X: 98.0, Y: 14.3}, {X: 96.0, Y: 13.3}, {X: 94.0, Y: 12.3}, {X: 91.0, Y: 12.3}, {X: 89.0, Y: 11.3}, {X: 88.0, Y: 11.3}, {X: 86.0, Y: 11.3}, {X: 85.0, Y: 11.3}, {X: 84.0, Y: 12.3}, {X: 83.0, Y: 13.3}, {X: 83.0, Y: 14.3},
			},
			// boucle for b
			{
				{X: 114.0, Y: 25.3}, {X: 114.0, Y: 24.3}, {X: 115.0, Y: 23.3}, {X: 116.0, Y: 22.3}, {X: 117.0, Y: 21.3}, {X: 118.0, Y: 21.3}, {X: 119.0, Y: 20.3}, {X: 120.0, Y: 20.3}, {X: 121.0, Y: 21.3}, {X: 122.0, Y: 21.3}, {X: 123.0, Y: 22.3}, {X: 124.0, Y: 22.3}, {X: 124.0, Y: 23.3}, {X: 125.0, Y: 24.3}, {X: 126.0, Y: 25.3}, {X: 126.0, Y: 26.3}, {X: 127.0, Y: 27.3}, {X: 127.0, Y: 29.3}, {X: 127.0, Y: 30.3}, {X: 127.0, Y: 32.3}, {X: 127.0, Y: 33.3}, {X: 127.0, Y: 34.3}, {X: 127.0, Y: 35.3}, {X: 127.0, Y: 36.3}, {X: 127.0, Y: 37.3}, {X: 126.0, Y: 38.3}, {X: 125.0, Y: 38.3}, {X: 124.0, Y: 38.3}, {X: 123.0, Y: 38.3}, {X: 122.0, Y: 38.3}, {X: 121.0, Y: 38.3}, {X: 120.0, Y: 38.3}, {X: 119.0, Y: 38.3}, {X: 118.0, Y: 37.3}, {X: 117.0, Y: 36.3}, {X: 117.0, Y: 35.3}, {X: 116.0, Y: 35.3}, {X: 116.0, Y: 34.3}, {X: 115.0, Y: 33.3},
			},
		},
	},

	// bezier then line
	{
		[]ShapeAtomKind{SAKBezier, SAKSegment},
		[]Shape{
			// l
			{
				{X: 87.0, Y: 20.3}, {X: 88.0, Y: 20.3}, {X: 89.0, Y: 20.3}, {X: 89.0, Y: 19.3}, {X: 90.0, Y: 19.3}, {X: 91.0, Y: 18.3}, {X: 91.0, Y: 17.3}, {X: 92.0, Y: 17.3}, {X: 92.0, Y: 16.3}, {X: 93.0, Y: 16.3}, {X: 93.0, Y: 15.3}, {X: 93.0, Y: 14.3}, {X: 94.0, Y: 14.3}, {X: 94.0, Y: 13.3}, {X: 94.0, Y: 12.3}, {X: 95.0, Y: 11.3}, {X: 95.0, Y: 10.3}, {X: 95.0, Y: 9.3}, {X: 95.0, Y: 8.3}, {X: 95.0, Y: 7.3}, {X: 95.0, Y: 6.3}, {X: 95.0, Y: 5.3}, {X: 95.0, Y: 4.3}, {X: 94.0, Y: 3.3}, {X: 93.0, Y: 3.3}, {X: 93.0, Y: 2.3}, {X: 92.0, Y: 2.3}, {X: 91.0, Y: 2.3}, {X: 90.0, Y: 2.3}, {X: 90.0, Y: 3.3}, {X: 90.0, Y: 4.3}, {X: 90.0, Y: 5.3}, {X: 89.0, Y: 6.3}, {X: 89.0, Y: 7.3}, {X: 89.0, Y: 8.3}, {X: 89.0, Y: 9.3}, {X: 89.0, Y: 10.3}, {X: 89.0, Y: 11.3}, {X: 89.0, Y: 12.3}, {X: 89.0, Y: 13.3}, {X: 89.0, Y: 14.3}, {X: 89.0, Y: 15.3}, {X: 89.0, Y: 16.3}, {X: 89.0, Y: 17.3}, {X: 89.0, Y: 18.3}, {X: 89.0, Y: 19.3}, {X: 89.0, Y: 20.3}, {X: 89.0, Y: 21.3}, {X: 89.0, Y: 22.3}, {X: 89.0, Y: 23.3}, {X: 89.0, Y: 24.3}, {X: 88.0, Y: 25.3}, {X: 88.0, Y: 26.3}, {X: 88.0, Y: 27.3}, {X: 88.0, Y: 28.3}, {X: 88.0, Y: 29.3}, {X: 88.0, Y: 30.3}, {X: 88.0, Y: 31.3}, {X: 88.0, Y: 32.3}, {X: 88.0, Y: 33.3}, {X: 89.0, Y: 33.3},
			},
		},
	},

	// circle then line
	{
		[]ShapeAtomKind{SAKCircle, SAKSegment},
		[]Shape{
			// a
			{
				{X: 114.0, Y: 19.3}, {X: 114.0, Y: 18.3}, {X: 113.0, Y: 18.3}, {X: 112.0, Y: 18.3}, {X: 111.0, Y: 17.3}, {X: 109.0, Y: 17.3}, {X: 108.0, Y: 17.3}, {X: 107.0, Y: 17.3}, {X: 106.0, Y: 18.3}, {X: 105.0, Y: 18.3}, {X: 104.0, Y: 19.3}, {X: 103.0, Y: 20.3}, {X: 103.0, Y: 21.3}, {X: 103.0, Y: 22.3}, {X: 103.0, Y: 23.3}, {X: 103.0, Y: 24.3}, {X: 103.0, Y: 25.3}, {X: 104.0, Y: 26.3}, {X: 104.0, Y: 27.3}, {X: 105.0, Y: 28.3}, {X: 106.0, Y: 29.3}, {X: 107.0, Y: 29.3}, {X: 108.0, Y: 29.3}, {X: 109.0, Y: 29.3}, {X: 111.0, Y: 28.3}, {X: 112.0, Y: 28.3}, {X: 112.0, Y: 27.3}, {X: 113.0, Y: 25.3}, {X: 114.0, Y: 24.3}, {X: 114.0, Y: 22.3}, {X: 115.0, Y: 21.3}, {X: 115.0, Y: 19.3}, {X: 115.0, Y: 18.3}, {X: 115.0, Y: 17.3}, {X: 115.0, Y: 16.3}, {X: 115.0, Y: 17.3}, {X: 115.0, Y: 18.3}, {X: 115.0, Y: 20.3}, {X: 115.0, Y: 21.3}, {X: 116.0, Y: 23.3}, {X: 116.0, Y: 25.3}, {X: 117.0, Y: 26.3}, {X: 117.0, Y: 27.3}, {X: 118.0, Y: 28.3}, {X: 118.0, Y: 29.3},
			},
		},
	},

	// line then bezier then Line
	{
		[]ShapeAtomKind{SAKSegment, SAKBezier, SAKSegment},
		[]Shape{
			// M
			{
				{X: 102.0, Y: 37.3}, {X: 103.0, Y: 37.3}, {X: 103.0, Y: 36.3}, {X: 103.0, Y: 35.3}, {X: 103.0, Y: 34.3}, {X: 103.0, Y: 33.3}, {X: 103.0, Y: 32.3}, {X: 103.0, Y: 30.3}, {X: 103.0, Y: 28.3}, {X: 103.0, Y: 27.3}, {X: 103.0, Y: 25.3}, {X: 103.0, Y: 23.3}, {X: 103.0, Y: 21.3}, {X: 103.0, Y: 20.3}, {X: 103.0, Y: 18.3}, {X: 103.0, Y: 16.3}, {X: 103.0, Y: 15.3}, {X: 103.0, Y: 14.3}, {X: 103.0, Y: 13.3}, {X: 102.0, Y: 12.3}, {X: 102.0, Y: 11.3}, {X: 103.0, Y: 11.3}, {X: 104.0, Y: 12.3}, {X: 105.0, Y: 12.3}, {X: 106.0, Y: 13.3}, {X: 107.0, Y: 13.3}, {X: 108.0, Y: 14.3}, {X: 110.0, Y: 15.3}, {X: 111.0, Y: 16.3}, {X: 112.0, Y: 17.3}, {X: 113.0, Y: 18.3}, {X: 115.0, Y: 19.3}, {X: 116.0, Y: 19.3}, {X: 117.0, Y: 20.3}, {X: 118.0, Y: 21.3}, {X: 119.0, Y: 21.3}, {X: 120.0, Y: 22.3}, {X: 121.0, Y: 22.3}, {X: 122.0, Y: 23.3}, {X: 122.0, Y: 22.3}, {X: 123.0, Y: 22.3}, {X: 124.0, Y: 21.3}, {X: 125.0, Y: 20.3}, {X: 126.0, Y: 18.3}, {X: 127.0, Y: 17.3}, {X: 128.0, Y: 16.3}, {X: 129.0, Y: 14.3}, {X: 130.0, Y: 12.3}, {X: 130.0, Y: 11.3}, {X: 131.0, Y: 9.3}, {X: 132.0, Y: 8.3}, {X: 132.0, Y: 7.3}, {X: 133.0, Y: 6.3}, {X: 133.0, Y: 7.3}, {X: 134.0, Y: 8.3}, {X: 134.0, Y: 9.3}, {X: 135.0, Y: 11.3}, {X: 135.0, Y: 12.3}, {X: 135.0, Y: 14.3}, {X: 136.0, Y: 16.3}, {X: 136.0, Y: 19.3}, {X: 136.0, Y: 21.3}, {X: 136.0, Y: 23.3}, {X: 137.0, Y: 25.3}, {X: 137.0, Y: 27.3}, {X: 137.0, Y: 29.3}, {X: 137.0, Y: 30.3}, {X: 138.0, Y: 32.3}, {X: 138.0, Y: 33.3}, {X: 138.0, Y: 35.3}, {X: 138.0, Y: 36.3}, {X: 138.0, Y: 37.3}, {X: 138.0, Y: 39.3}, {X: 138.0, Y: 38.3}, {X: 137.0, Y: 38.3},
			},
		},
	},

	// four lines
	{
		[]ShapeAtomKind{SAKSegment, SAKSegment, SAKSegment, SAKSegment},
		[]Shape{
			{
				{X: 114.0, Y: 6.3}, {X: 112.0, Y: 7.3}, {X: 111.0, Y: 8.3}, {X: 109.0, Y: 9.3}, {X: 106.0, Y: 10.3}, {X: 104.0, Y: 11.3}, {X: 102.0, Y: 12.3}, {X: 100.0, Y: 13.3}, {X: 98.0, Y: 14.3}, {X: 97.0, Y: 14.3}, {X: 95.0, Y: 15.3}, {X: 94.0, Y: 16.3}, {X: 93.0, Y: 16.3}, {X: 92.0, Y: 16.3}, {X: 93.0, Y: 17.3}, {X: 94.0, Y: 17.3}, {X: 95.0, Y: 17.3}, {X: 96.0, Y: 18.3}, {X: 97.0, Y: 18.3}, {X: 98.0, Y: 18.3}, {X: 100.0, Y: 19.3}, {X: 101.0, Y: 19.3}, {X: 102.0, Y: 20.3}, {X: 104.0, Y: 20.3}, {X: 105.0, Y: 20.3}, {X: 106.0, Y: 21.3}, {X: 107.0, Y: 21.3}, {X: 108.0, Y: 21.3}, {X: 109.0, Y: 22.3}, {X: 110.0, Y: 22.3}, {X: 111.0, Y: 22.3}, {X: 112.0, Y: 23.3}, {X: 113.0, Y: 23.3}, {X: 114.0, Y: 24.3}, {X: 115.0, Y: 24.3}, {X: 115.0, Y: 25.3}, {X: 114.0, Y: 26.3}, {X: 113.0, Y: 27.3}, {X: 112.0, Y: 28.3}, {X: 110.0, Y: 29.3}, {X: 109.0, Y: 30.3}, {X: 107.0, Y: 31.3}, {X: 106.0, Y: 32.3}, {X: 105.0, Y: 34.3}, {X: 103.0, Y: 35.3}, {X: 102.0, Y: 35.3}, {X: 101.0, Y: 36.3}, {X: 100.0, Y: 37.3}, {X: 99.0, Y: 38.3}, {X: 98.0, Y: 38.3}, {X: 98.0, Y: 39.3}, {X: 99.0, Y: 39.3}, {X: 101.0, Y: 39.3}, {X: 102.0, Y: 39.3}, {X: 104.0, Y: 40.3}, {X: 106.0, Y: 40.3}, {X: 108.0, Y: 40.3}, {X: 111.0, Y: 40.3}, {X: 113.0, Y: 40.3}, {X: 115.0, Y: 39.3}, {X: 117.0, Y: 39.3}, {X: 119.0, Y: 39.3}, {X: 120.0, Y: 38.3}, {X: 121.0, Y: 38.3}, {X: 122.0, Y: 38.3}, {X: 123.0, Y: 38.3}, {X: 125.0, Y: 37.3}, {X: 125.0, Y: 36.3}, {X: 126.0, Y: 35.3},
			},
		},
	},
}

func assertKinds(t *testing.T, got []ShapeAtom, expected []ShapeAtomKind) {
	t.Helper()

	tu.Assert(t, len(got) == len(expected))
	for i := range got {
		tu.Assert(t, got[i].Kind() == expected[i])
	}
}

func TestRealSample(t *testing.T) {
	for _, v := range realShapes {
		for _, s := range v.inputs {
			assertKinds(t, s.SubShapes().Identify(), v.expected)
		}
	}
}

func TestNoisyCircle(t *testing.T) {
	// o with missing points
	noisyO := Shape{
		{X: 24.0, Y: 18.3}, {X: 23.0, Y: 17.3}, {X: 22.0, Y: 17.3}, {X: 21.0, Y: 17.3}, {X: 20.0, Y: 17.3}, {X: 19.0, Y: 17.3}, {X: 18.0, Y: 17.3}, {X: 18.0, Y: 18.3}, {X: 17.0, Y: 18.3}, {X: 17.0, Y: 20.3}, {X: 16.0, Y: 21.3}, {X: 16.0, Y: 22.3}, {X: 16.0, Y: 24.3}, {X: 16.0, Y: 25.3}, {X: 17.0, Y: 27.3}, {X: 17.0, Y: 28.3}, {X: 18.0, Y: 29.3}, {X: 19.0, Y: 29.3}, {X: 20.0, Y: 29.3}, {X: 21.0, Y: 29.3}, {X: 22.0, Y: 29.3}, {X: 23.0, Y: 28.3}, {X: 23.0, Y: 26.3}, {X: 24.0, Y: 25.3}, {X: 24.0, Y: 23.3}, {X: 24.0, Y: 21.3}, {X: 24.0, Y: 19.3}, {X: 24.0, Y: 17.3}, {X: 24.0, Y: 16.3},
	}

	tu.Assert(t, noisyO.identify().Kind() == SAKCircle)
}

func TestVerticalLine(t *testing.T) {
	vert := Shape{
		{X: 25.0, Y: 24.3}, {X: 25.0, Y: 23.3}, {X: 25.0, Y: 22.3}, {X: 25.0, Y: 21.3}, {X: 25.0, Y: 19.3}, {X: 25.0, Y: 18.3},
	}
	tu.Assert(t, vert.identify().Kind() == SAKSegment)
}
