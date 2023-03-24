package symbols

import (
	"os"
	"path/filepath"
	"testing"

	tu "github.com/benoitkugler/pen2latex/testutils"
)

func TestSerialize(t *testing.T) {
	base := map[rune]Symbol{}
	for _, group := range symbols {
		r := rune(group.description[0])
		base[r] = group.symbols[0]
	}
	db := NewStore(base)
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, "database.pen2latex")
	err := db.Serialize(path)
	tu.AssertNoErr(t, err)

	db2, err := NewStoreFromDisk(path)
	tu.AssertNoErr(t, err)
	tu.AssertEqual(t, len(db2.Symbols), len(db.Symbols))
}
