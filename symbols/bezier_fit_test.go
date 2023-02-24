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
)

func TestPathIndices(t *testing.T) {
	points := []Pos{
		{0, 0}, {1, 1}, {1, 1}, {1, 8},
	}
	indices := pathLengthIndices(points)
	if indices[0] != 0 || indices[len(indices)-1] != 1 {
		t.Fatal()
	}
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

func generateBezierPoints(b BezierC) Shape {
	var sh Shape
	for t := range [20]int{} {
		sh = append(sh, b.eval(fl(t)/20))
	}
	return sh
}

func printShape(t *testing.T, sh Shape, filename string) Shape {
	graph := graph2D(sh)
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, graph)
	if err != nil {
		t.Fatal(err)
	}

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

func TestFitBezier(t *testing.T) {
	tmpDir := os.TempDir()

	origin := BezierC{Pos{}, Pos{30, 40}, Pos{50, -40}, Pos{60, 0}}
	points := generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube1_origin.png"))

	fitted, _ := fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube1_fitted.png"))

	origin = BezierC{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	points = generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube2_origin.png"))

	fitted, _ = fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube2_fitted.png"))

	// linear shape
	origin = BezierC{Pos{}, Pos{30, 30}, Pos{40, 40}, Pos{60, 60}}
	points = generateBezierPoints(origin)
	printShape(t, points, filepath.Join(tmpDir, "bezier_cube3_origin.png"))

	fitted, _ = fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube3_fitted.png"))

	_, errLine := fitSegment(points)
	if errLine != 0 {
		t.Fatal()
	}

	// data from a circle, not a bezier !
	points = circle(Pos{30, 30}, 20)
	fitted, errCube := fitCubicBezier(points)
	printShape(t, generateBezierPoints(fitted), filepath.Join(tmpDir, "bezier_cube4_fitted.png"))
	_, errCircle := fitCircle(points)
	if errCircle > errCube {
		t.Fatal()
	}
}

func TestIdentify(t *testing.T) {
	origin := BezierC{Pos{}, Pos{0, 40}, Pos{50, 40}, Pos{60, 0}}
	points := generateBezierPoints(origin)
	if points.identify().Kind() != SAKBezier {
		t.Fatal()
	}

	points = circle(Pos{30, 30}, 20)
	if points.identify().Kind() != SAKCircle {
		t.Fatal()
	}
}
