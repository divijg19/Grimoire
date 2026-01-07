package adapters

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/divijg19/Grimoire/internal/engine"
	"github.com/divijg19/Grimoire/internal/ports"
)

// JSONStore implements ports.Store using a JSON file.
type JSONStore struct {
	Path string
}

// NewJSONStore creates a JSON-backed store at the given path.
func NewJSONStore(path string) ports.Store {
	return &JSONStore{Path: path}
}

// Load loads the game state or returns DefaultState if missing/corrupt.
func (s *JSONStore) Load() (*engine.State, error) {
	if _, err := os.Stat(s.Path); errors.Is(err, os.ErrNotExist) {
		state := engine.DefaultState()
		return &state, nil
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		state := engine.DefaultState()
		return &state, err
	}

	var state engine.State
	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupt save: move aside
		ts := time.Now().Unix()
		corrupt := s.Path + ".corrupt." + intToString(ts)
		_ = os.Rename(s.Path, corrupt)

		def := engine.DefaultState()
		return &def, err
	}

	// Ensure inventory map exists
	state.Player.EnsureInventory()
	state.Player.ClampHP()

	return &state, nil
}

// Save writes the state atomically.
func (s *JSONStore) Save(state *engine.State) error {
	tmp := s.Path + ".tmp"

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	// Atomic replace
	return os.Rename(tmp, s.Path)
}

// ================================
// Helpers
// ================================

func intToString(v int64) string {
	return filepath.Base(time.Unix(v, 0).Format("20060102_150405"))
}
