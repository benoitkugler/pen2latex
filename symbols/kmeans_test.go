package symbols

import (
	"reflect"
	"testing"
)

func TestKmeans(t *testing.T) {
	// 'clear' two cluster case
	input := []fl{
		100, 99, 98, 100, 0, 0, -0.1, 0.1,
	}
	clusters := segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 4}, {4, 8}}) {
		t.Fatal(clusters)
	}

	// two cluster with outlier
	input = []fl{
		100, 100, 101, 101, 50, 0, 0, -0.1, 0.2,
	}
	clusters = segmentation(input)
	if len(clusters) != 2 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 4}, {5, 9}}) {
		t.Fatal(clusters)
	}

	// linear form
	input = []fl{
		10, 11, 12, 13, 14, 15,
	}
	clusters = segmentation(input)
	if len(clusters) != 1 {
		t.Fatal()
	}
	if !reflect.DeepEqual(clusters, []clusterRange{{0, 6}}) {
		t.Fatal()
	}
}
