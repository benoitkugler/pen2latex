package symbols

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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

// SymbolStore stores the shape of
// runes, as setup by the user,
// and is later used to map a mouse entry to a rune.
type SymbolStore struct {
	// acts as a map[rune][]Symbol, but with faster iteration,
	entries []mapEntry
}

// NewSymbolStore return a database for the given [shapes].
func NewSymbolStore(shapes map[rune]Symbol) *SymbolStore {
	entries := make([]mapEntry, 0, len(shapes))
	for r, sy := range shapes {
		entries = append(entries, mapEntry{sy.SegmentToAtoms(), r})
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

	for _, ee := range out.entries {
		if ee.R == '2' || ee.R == '3' {
			fmt.Println(string(ee.R), ee.Shape)
		}
	}
	return &out, nil
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
	Shape ShapeFootprint `json:"s"`
	R     rune           `json:"r"`
}

type ShapeFootprint []ShapeAtom

func (sf ShapeFootprint) String() string {
	chunks := make([]string, len(sf))
	for i, a := range sf {
		chunks[i] = fmt.Sprintf("%s%v", a.Kind(), a)
	}
	return "[ " + strings.Join(chunks, " ; ") + " ]"
}

func (l ShapeFootprint) MarshalJSON() ([]byte, error) {
	tmp := make([]shapeAtomData, len(l))
	for i, a := range l {
		tmp[i] = a.serialize()
	}
	return json.Marshal(tmp)
}

func (l *ShapeFootprint) UnmarshalJSON(data []byte) error {
	var tmp []shapeAtomData
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*l = make(ShapeFootprint, len(tmp))
	for i, d := range tmp {
		(*l)[i], err = d.deserialize()
		if err != nil {
			return err
		}
	}
	return nil
}
