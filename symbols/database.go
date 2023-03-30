package symbols

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Store stores the shapes of
// runes, as setup by the user,
// and is later used to map a mouse entry to a rune.
type Store struct {
	// Symbols acts as a map[rune][]Symbols, but with faster iteration,
	Symbols []RuneFootprint
}

// NewStore return a database for the given [symbols].
func NewStore(symbols map[rune]Symbol) Store {
	out := Store{
		Symbols: make([]RuneFootprint, 0, len(symbols)),
	}
	for r, sy := range symbols {
		fp := sy.Footprint()
		out.Symbols = append(out.Symbols, RuneFootprint{fp, r})
	}
	sort.Slice(out.Symbols, func(i, j int) bool { return out.Symbols[i].R < out.Symbols[j].R })

	return out
}

// NewStoreFromDisk load a store previously saved with
// [Serialize]
func NewStoreFromDisk(filename string) (Store, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Store{}, fmt.Errorf("opening on-disk store: %s", err)
	}
	defer f.Close()

	var out Store
	err = json.NewDecoder(f).Decode(&out)
	if err != nil {
		return Store{}, fmt.Errorf("deserializing on-disk store: %s", err)
	}

	return out, nil
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

type RuneFootprint struct {
	Footprint Footprint `json:"s"`
	R         rune      `json:"r"`
}

func (st Store) MarshalJSON() ([]byte, error) { return json.Marshal(st.Symbols) }

func (st *Store) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, &st.Symbols) }
