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
	EMWidth         float32 = 20.
	EMHeight        float32 = 50.
	EMBaselineRatio float32 = 0.66 // from the top
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
		entries = append(entries, mapEntry{sy.Union().normalizeX(), r})
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
//
// [boundingBox] is the parent scope, independent on the symbol,
// and used to normalize the given symbol to match the database reference geometry.
//
// It will panic is the store is empty.
func (ss SymbolStore) Lookup(rec Record, boundingBox Rect) (r rune, preferCompound bool) {
	return ss.match(rec)
	// var (
	// 	bestIndexCompound, bestIndexShape int
	// 	bestDistCompound, bestDistShape   float32 = math.MaxFloat32, math.MaxFloat32
	// )
	// fmt.Println(boundingBox.IsEmpty())
	// compoundShape := rec.Compound().Union().normalizeX().normalizeY(boundingBox)
	// shape := rec.Shape().normalizeX().normalizeY(boundingBox)

	// for i, entry := range ss.entries {
	// 	if score := frechetDistanceShapes(compoundShape, entry.Shape); score < bestDistCompound {
	// 		bestIndexCompound = i
	// 		bestDistCompound = score
	// 	}
	// 	if score := frechetDistanceShapes(shape, entry.Shape); score < bestDistShape {
	// 		bestIndexShape = i
	// 		bestDistShape = score
	// 	}
	// }

	// // If shape is adjacent to compound, always prefer compound
	// // do not normalize !
	// if closestPointDistance(rec.Shape(), rec.LastCompound().Union()) < penWidth {
	// 	return ss.entries[bestIndexCompound].R, true
	// }

	// if bestDistCompound < bestDistShape {
	// 	return ss.entries[bestIndexCompound].R, true
	// }
	// return ss.entries[bestIndexShape].R, false
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
