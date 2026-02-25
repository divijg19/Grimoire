package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONStoreLoad_MissingFileReturnsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	store := &JSONStore{Path: path}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if state.Player.Name != "Traveller" {
		t.Fatalf("expected default state, got player name %q", state.Player.Name)
	}
}

func TestJSONStoreLoad_CorruptFileRenamed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write corrupt payload: %v", err)
	}

	store := &JSONStore{Path: path}
	state, err := store.Load()
	if err == nil {
		t.Fatalf("expected unmarshal error for corrupt save")
	}
	if state.Player.Name != "Traveller" {
		t.Fatalf("expected default state fallback, got player name %q", state.Player.Name)
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("read dir: %v", readErr)
	}

	foundCorrupt := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "save.json.corrupt.") {
			foundCorrupt = true
			break
		}
	}
	if !foundCorrupt {
		t.Fatalf("expected renamed corrupt backup file")
	}
}

func TestJSONStoreSave_ThenLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")
	store := &JSONStore{Path: path}

	state, err := store.Load()
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}
	state.Player.Gold = 777
	state.Player.Inventory["healing_potion"] = 3

	if err := store.Save(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if reloaded.Player.Gold != 777 {
		t.Fatalf("expected persisted gold 777, got %d", reloaded.Player.Gold)
	}
	if reloaded.Player.Inventory["healing_potion"] != 3 {
		t.Fatalf("expected persisted healing_potion count 3, got %d", reloaded.Player.Inventory["healing_potion"])
	}
}
