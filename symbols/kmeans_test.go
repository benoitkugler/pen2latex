package symbols

import (
	"reflect"
	"testing"
)

func TestKmeans(t *testing.T) {
	// 'clear' two cluster case
	input := []Pos{
		{1, 100}, {2, 100}, {3, 101}, {4, 0}, {5, -0.1}, {6, 0.2},
	}
	clusters := segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 3}, {3, 6}}) {
		t.Fatal(clusters)
	}

	// two cluster with outlier
	input = []Pos{
		{1, 100}, {2, 100}, {3, 101}, {4, 50}, {5, 0}, {6, -0.1}, {7, 0.2},
	}
	clusters = segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 3}, {4, 7}}) {
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
