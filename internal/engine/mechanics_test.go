package engine

import "testing"

type seqRNG struct {
	ints   []int
	floats []float64
	i      int
	f      int
}

func (r *seqRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if r.i >= len(r.ints) {
		return 0
	}
	v := r.ints[r.i]
	r.i++
	if v < 0 {
		v = -v
	}
	return v % n
}

func (r *seqRNG) Float64() float64 {
	if r.f >= len(r.floats) {
		return 0
	}
	v := r.floats[r.f]
	r.f++
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func TestRest_SpendsSPAndClampsHP(t *testing.T) {
	state := DefaultState()
	state.Player.HP = 80
	state.Player.SP = 3

	events, err := Rest(&state, 2)
	if err != nil {
		t.Fatalf("Rest returned error: %v", err)
	}
	if state.Player.SP != 1 {
		t.Fatalf("expected SP=1, got %d", state.Player.SP)
	}
	if state.Player.HP != state.Player.MaxHP {
		t.Fatalf("expected HP clamped to max=%d, got %d", state.Player.MaxHP, state.Player.HP)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestRest_NotEnoughSP(t *testing.T) {
	state := DefaultState()
	state.Player.SP = 0

	_, err := Rest(&state, 1)
	if err == nil {
		t.Fatalf("expected error when SP is insufficient")
	}
}

func TestGrantXP_MultipleLevelUps(t *testing.T) {
	state := DefaultState()
	state.Player.Level = 1
	state.Player.XP = 0
	state.Player.MaxHP = 100
	state.Player.HP = 90

	events := GrantXP(&state, 350)
	if len(events) < 3 {
		t.Fatalf("expected xp + level events, got %d", len(events))
	}
	if state.Player.Level != 3 {
		t.Fatalf("expected level 3, got %d", state.Player.Level)
	}
	if state.Player.XP != 50 {
		t.Fatalf("expected remaining XP 50, got %d", state.Player.XP)
	}
	if state.Player.MaxHP != 120 {
		t.Fatalf("expected MaxHP 120, got %d", state.Player.MaxHP)
	}
	if state.Player.HP != 110 {
		t.Fatalf("expected HP 110 after level-up heals, got %d", state.Player.HP)
	}
}

func TestResolveCombat_WinDeterministicWithLoot(t *testing.T) {
	state := DefaultState()
	state.Player.Level = 10
	state.Player.HP = 100

	rng := &seqRNG{ints: []int{0}, floats: []float64{0.05, 0.05}}
	result, events := ResolveCombat(&state, Enemies["goblin"], rng)

	if result.Outcome != "win" {
		t.Fatalf("expected win, got %q", result.Outcome)
	}
	if result.XP != Enemies["goblin"].XP || result.Gold != Enemies["goblin"].Gold {
		t.Fatalf("unexpected rewards: xp=%d gold=%d", result.XP, result.Gold)
	}
	if len(result.Loot) == 0 {
		t.Fatalf("expected deterministic loot drop")
	}
	if len(events) < 2 {
		t.Fatalf("expected encounter and defeat events")
	}
	if state.Player.HP != 100 {
		t.Fatalf("expected no HP loss, got %d", state.Player.HP)
	}
}

func TestHunt_UsesExtraSPAndAppliesMultiplier(t *testing.T) {
	state := DefaultState()
	state.Player.Level = 10
	state.Player.XP = 0
	state.Player.SP = 10
	state.Player.Gold = 0

	rng := &seqRNG{ints: []int{0, 0}, floats: []float64{1, 1}}
	events, err := Hunt(&state, 2, rng)
	if err != nil {
		t.Fatalf("Hunt returned error: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected hunt events")
	}

	if state.Player.SP != 7 {
		t.Fatalf("expected SP 7 after spending 3, got %d", state.Player.SP)
	}
	if state.Player.XP != 7 {
		t.Fatalf("expected XP 7 (5 * 1.5), got %d", state.Player.XP)
	}
	if state.Player.Gold != 4 {
		t.Fatalf("expected Gold 4 (3 * 1.5), got %d", state.Player.Gold)
	}
}

func TestHunt_ClampsExtraSPToMax(t *testing.T) {
	state := DefaultState()
	state.Player.Level = 10
	state.Player.SP = 20

	rng := &seqRNG{ints: []int{0, 0}, floats: []float64{1, 1}}
	_, err := Hunt(&state, 99, rng)
	if err != nil {
		t.Fatalf("Hunt returned error: %v", err)
	}

	wantSpent := HuntBaseSP + HuntExtraSPMax
	if state.Player.SP != 20-wantSpent {
		t.Fatalf("expected SP %d, got %d", 20-wantSpent, state.Player.SP)
	}
}

func TestExplore_TreasurePath(t *testing.T) {
	state := DefaultState()
	state.Player.Gold = 0

	rng := &seqRNG{ints: []int{0, 0, 1}}
	events, err := Explore(&state, rng)
	if err != nil {
		t.Fatalf("Explore returned error: %v", err)
	}

	if state.Meta.CommandCount != 1 {
		t.Fatalf("expected command count incremented, got %d", state.Meta.CommandCount)
	}
	if state.Player.Gold != 100 {
		t.Fatalf("expected 100 gold from deterministic treasure roll, got %d", state.Player.Gold)
	}
	if !HasItem(&state.Player, "rusty_dagger", 2) {
		t.Fatalf("expected rusty_dagger incremented from treasure item roll")
	}
	if len(events) < 3 {
		t.Fatalf("expected treasure, gold, and item events; got %d", len(events))
	}
}

func TestExplore_PlayerDownReturnsError(t *testing.T) {
	state := DefaultState()
	state.Player.HP = 0

	_, err := Explore(&state, &seqRNG{})
	if err == nil {
		t.Fatalf("expected error when player HP is 0")
	}
}
