package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONStoreLoad_NormalizesInventoryIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "save.json")

	payload := `{
  "player": {
    "name": "Traveller",
    "class": "Adventurer",
    "gold": 50,
    "hp": 100,
    "max_hp": 100,
    "sp": 10,
    "level": 1,
    "xp": 0,
    "inventory": {
      "rusty dagger": 1,
      "rusty_dagger": 2,
      "Torch": 1,
      "ignored": 0
    }
  },
  "meta": {
    "location": "Starting Village",
    "quests_completed": 0,
    "command_count": 0
  }
}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	store := &JSONStore{Path: path}
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got := state.Player.Inventory["rusty_dagger"]; got != 3 {
		t.Fatalf("expected merged rusty_dagger=3, got %d", got)
	}
	if got := state.Player.Inventory["torch"]; got != 1 {
		t.Fatalf("expected normalized torch=1, got %d", got)
	}
	if _, ok := state.Player.Inventory["rusty dagger"]; ok {
		t.Fatalf("unexpected non-canonical key present")
	}
}
