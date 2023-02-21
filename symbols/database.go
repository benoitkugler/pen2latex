package symbols

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
)

// var RequiredRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var RequiredRunes = []rune("abcdefxy()")

const (
	penWidth = 4
)

// SymbolStore stores the shape of
// runes, as setup by the user,
// and is later used to map a mouse entry to a rune.
type SymbolStore struct {
	entries []mapEntry // acts as a map, but with faster iteration
}

// NewSymbolStore return a database for the given [shapes].
func NewSymbolStore(shapes map[rune]Symbol) *SymbolStore {
	entries := make([]mapEntry, 0, len(shapes))
	for r, sy := range shapes {
		entries = append(entries, mapEntry{sy.Union().normalized(), r})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].R < entries[j].R })
	return &SymbolStore{entries: entries}
}

// NewSymbolStoreFromDisk load a store previously saved with
// [Serialize]
func NewSymbolStoreFromDisk(filename string) (*SymbolStore, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening on-disk store: %s", err)
	}
	defer f.Close()

	var out SymbolStore
	err = json.NewDecoder(f).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("deserializing on-disk store: %s", err)
	}
	return &out, nil
}

// [Lookup] performs approximate matching by finding
// the closest shape in the database and returning its rune.
// More precisely, it compares scores for [rec.Shape()] and [rec.Compound()]
// returning which is better in [preferCompound].
// It will panic is the store is empty.
func (ss SymbolStore) Lookup(rec Record) (r rune, preferCompound bool) {
	var (
		bestIndexCompound, bestIndexShape int
		bestDistCompound, bestDistShape   float32 = math.MaxFloat32, math.MaxFloat32
	)
	compoundShape := rec.Compound().Union().normalized()
	shape := rec.Shape().normalized()

	for i, entry := range ss.entries {
		if score := frechetDistanceShapes(compoundShape, entry.Shape); score < bestDistCompound {
			bestIndexCompound = i
			bestDistCompound = score
		}
		if score := frechetDistanceShapes(shape, entry.Shape); score < bestDistShape {
			bestIndexShape = i
			bestDistShape = score
		}
	}

	// If shape is adjacent to compound, always prefer compound
	if closestPointDistance(shape, rec.LastCompound().Union().normalized()) < penWidth {
		return ss.entries[bestIndexCompound].R, true
	}

	if bestDistCompound < bestDistShape {
		return ss.entries[bestIndexCompound].R, true
	}
	return ss.entries[bestIndexShape].R, false
}

// Serialize dumps the store into [filename]
func (ss SymbolStore) Serialize(filename string) error {
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

func (ss SymbolStore) MarshalJSON() ([]byte, error) {
	return json.Marshal(ss.entries)
}

func (ss *SymbolStore) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &ss.entries)
}

type mapEntry struct {
	Shape Shape `json:"s"`
	R     rune  `json:"r"`
}
