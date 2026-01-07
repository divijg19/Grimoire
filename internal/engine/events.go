package engine

// ================================
// Event System
// ================================

// Event is a marker interface.
// Engine logic emits events; UI decides how to render them.
type Event interface {
	EventType() string
}

// ================================
// Core Events
// ================================

// DamageDealt is emitted when damage occurs in combat.
type DamageDealt struct {
	Source string // "player" or enemy ID
	Target string // "player" or enemy ID
	Amount int
	HPLeft int
}

func (DamageDealt) EventType() string { return "damage_dealt" }

// EnemyDefeated is emitted when an enemy dies.
type EnemyDefeated struct {
	EnemyID string
	XP      int
	Gold    int
}

func (EnemyDefeated) EventType() string { return "enemy_defeated" }

// PlayerDefeated is emitted when the player reaches 0 HP.
type PlayerDefeated struct{}

func (PlayerDefeated) EventType() string { return "player_defeated" }

// ================================
// Progression Events
// ================================

// XPGained is emitted when XP is awarded.
type XPGained struct {
	Amount int
}

func (XPGained) EventType() string { return "xp_gained" }

// LevelUp is emitted when the player levels up.
type LevelUp struct {
	NewLevel int
	NewMaxHP int
}

func (LevelUp) EventType() string { return "level_up" }

// ================================
// Inventory & Loot Events
// ================================

// ItemAdded is emitted when an item enters inventory.
type ItemAdded struct {
	ItemID string
	Count  int
}

func (ItemAdded) EventType() string { return "item_added" }

// ItemRemoved is emitted when an item is consumed or lost.
type ItemRemoved struct {
	ItemID string
	Count  int
}

func (ItemRemoved) EventType() string { return "item_removed" }

// LootFound is emitted after combat or exploration.
type LootFound struct {
	Items []string
}

func (LootFound) EventType() string { return "loot_found" }

// ================================
// Resource Events
// ================================

// GoldGained is emitted when gold is gained.
type GoldGained struct {
	Amount int
}

func (GoldGained) EventType() string { return "gold_gained" }

// SPSpent is emitted when SP is consumed.
type SPSpent struct {
	Amount int
}

func (SPSpent) EventType() string { return "sp_spent" }

// HPRestored is emitted when HP is restored.
type HPRestored struct {
	Amount int
}

func (HPRestored) EventType() string { return "hp_restored" }

// ================================
// World / Flow Events
// ================================

// EncounterStarted signals an enemy encounter.
type EncounterStarted struct {
	EnemyID string
}

func (EncounterStarted) EventType() string { return "encounter_started" }

// ExplorationResult is emitted for non-combat explore outcomes.
type ExplorationResult struct {
	Kind string // "nothing", "gold", "item", "treasure"
}

func (ExplorationResult) EventType() string { return "exploration_result" }

// ================================
// Utility
// ================================

// Events is a convenience alias.
type Events []Event
