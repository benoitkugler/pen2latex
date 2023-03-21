package symbols

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

// shared functions for tests

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
	return image.Rect(int(minX*8)-2, int(minY*8)-2, int(maxX*8)+2, int(maxY*8)+2)
}

// return a rough version of a graph (sh[i].X, sh[i].Y)
func graph2D(sh Shape) image.Image {
	// adjust the graph height
	rect := minMax2D(sh)
	out := image.NewGray(rect)
	for _, v := range sh {
		out.SetGray(int(v.X*8), int(v.Y*8), color.Gray{255})
	}
	return out
}

func printShape(t *testing.T, sh Shape, filename string) {
	t.Helper()

	dir := os.TempDir()
	filename = filepath.Join(dir, filename+".png")

	graph := graph2D(sh)
	f, err := os.Create(filename)
	tu.AssertNoErr(t, err)

	err = png.Encode(f, graph)
	tu.AssertNoErr(t, err)

	fmt.Printf("Graph saved in file://%s\n", filename)
}
