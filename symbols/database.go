package symbols

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// var RequiredRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var RequiredRunes = []rune("abcdefxy()123")

const (
	penWidth = 4
)

const (
	EMWidth         float32 = 30.
	EMHeight        float32 = 60.
	EMBaselineRatio float32 = 0.66 // from the top
)

// Store stores the shapes of
// runes, as setup by the user,
// and is later used to map a mouse entry to a rune.
type Store struct {
	// acts as a map[rune][]Symbol, but with faster iteration,
	entries []mapEntry
}

// NewStore return a database for the given [symbols].
func NewStore(symbols map[rune]Symbol) *Store {
	out := &Store{
		entries: make([]mapEntry, 0, len(symbols)),
	}
	for r, sy := range symbols {
		fp := sy.Footprint()
		out.entries = append(out.entries, mapEntry{fp, r})
	}
	sort.Slice(out.entries, func(i, j int) bool { return out.entries[i].R < out.entries[j].R })

	return out
}

// NewStoreFromDisk load a store previously saved with
// [Serialize]
func NewStoreFromDisk(filename string) (*Store, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening on-disk store: %s", err)
	}
	defer f.Close()

	var out Store
	err = json.NewDecoder(f).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("deserializing on-disk store: %s", err)
	}

	return &out, nil
}

// Serialize dumps the store into [filename]
func (ss Store) Serialize(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating on-disk store: %s", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent(" ", " ")
	err = enc.Encode(ss)
	if err != nil {
		return fmt.Errorf("serializing on-disk store: %s", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("closing on-disk store: %s", err)
	}
	return nil
}

type mapEntry struct {
	Footprint SymbolFootprint `json:"s"`
	R         rune            `json:"r"`
}
