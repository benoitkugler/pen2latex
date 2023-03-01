package symbols

import (
	"reflect"
	"testing"
)

func TestKmeans(t *testing.T) {
	// 'clear' two cluster case
	input := []Pos{
		{1.5, 100}, {2, 100}, {2.5, 101}, {3, 100}, {4.5, 0}, {4.5, 0}, {4.5, -0.1}, {4.5, 0.1},
	}
	clusters := segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 4}, {4, 8}}) {
		t.Fatal(clusters)
	}

	// two cluster with outlier
	input = []Pos{
		{2, 100}, {2, 100}, {3, 101}, {3, 101}, {4, 50}, {5, 0}, {5, 0}, {6, -0.1}, {7, 0.2},
	}
	clusters = segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 4}, {5, 9}}) {
		t.Fatal(clusters)
	}

	// linear form
	input = []Pos{
		{1, 10}, {2, 11}, {3, 12}, {4, 13}, {5, 14}, {6, 15},
	}
	clusters = segmentation(input)
	if len(clusters) != 1 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 6}}) {
		t.Fatal()
	}
}
