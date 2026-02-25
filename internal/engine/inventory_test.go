package engine

import "testing"

type fixedRNG struct {
	ints []int
	idx  int
}

func (r *fixedRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if r.idx >= len(r.ints) {
		return 0
	}
	v := r.ints[r.idx]
	r.idx++
	if v < 0 {
		v = -v
	}
	return v % n
}

func (r *fixedRNG) Float64() float64 { return 0 }

func TestNormalizeItemID(t *testing.T) {
	tests := map[string]string{
		" Healing Potion ": "healing_potion",
		"RUSTY-DAGGER":     "rusty_dagger",
		"silver   ring":    "silver_ring",
		"":                 "",
	}

	for in, want := range tests {
		got := NormalizeItemID(in)
		if got != want {
			t.Fatalf("NormalizeItemID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeInventory_DeduplicatesAndFilters(t *testing.T) {
	inv := map[string]int{
		"rusty dagger": 1,
		"rusty_dagger": 2,
		"Torch":        1,
		"":             5,
		"bad":          0,
	}

	n := NormalizeInventory(inv)
	if n["rusty_dagger"] != 3 {
		t.Fatalf("expected rusty_dagger merged count 3, got %d", n["rusty_dagger"])
	}
	if n["torch"] != 1 {
		t.Fatalf("expected torch count 1, got %d", n["torch"])
	}
	if _, ok := n[""]; ok {
		t.Fatalf("expected empty key removed")
	}
	if _, ok := n["bad"]; ok {
		t.Fatalf("expected zero-count entry removed")
	}
}

func TestUseItem_AcceptsNonCanonicalID(t *testing.T) {
	state := DefaultState()
	AddItem(&state.Player, "healing_potion", 1)

	rng := &fixedRNG{ints: []int{0, 0}}
	events, err := UseItem(&state, "Healing Potion", rng)
	if err != nil {
		t.Fatalf("UseItem returned error: %v", err)
	}
	if HasItem(&state.Player, "healing_potion", 1) {
		t.Fatalf("expected potion consumed")
	}

	if len(events) == 0 {
		t.Fatalf("expected events emitted")
	}
}
