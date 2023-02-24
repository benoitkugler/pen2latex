// Package layout implements a data structure
// representing a math expression and its layout,
// used to recognize and correctly insert a user input
package layout

import "testing"

func TestNode_insertAt(t *testing.T) {
	tests := []struct {
		fields []block
		bl     block
		index  int
	}{
		{nil, &regularChar{}, 0},
	}
	for _, tt := range tests {
		n := &Node{
			Blocks: tt.fields,
		}
		n.insertAt(tt.bl, tt.index)
	}
}
