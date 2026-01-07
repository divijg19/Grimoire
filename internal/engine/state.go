package engine

// ================================
// Core State Definitions
// ================================

// State is the root game state.
// It is the ONLY mutable structure passed through engine rules.
type State struct {
	Player Player `json:"player"`
	Meta   Meta   `json:"meta"`
}

// ================================
// Player
// ================================

type Player struct {
	Name      string         `json:"name"`
	Class     string         `json:"class"`
	Gold      int            `json:"gold"`
	HP        int            `json:"hp"`
	MaxHP     int            `json:"max_hp"`
	SP        int            `json:"sp"`
	Level     int            `json:"level"`
	XP        int            `json:"xp"`
	Inventory map[string]int `json:"inventory"` // item_id -> count
}

// ================================
// Meta
// ================================

type Meta struct {
	Location        string `json:"location"`
	QuestsCompleted int    `json:"quests_completed"`
	CommandCount    int    `json:"command_count"`
}

// ================================
// Defaults
// ================================

const (
	DefaultMaxHP = 100

	HuntBaseSP      = 1
	HuntExtraSPMax  = 5
	RestHPPerSP     = 25
	RestockInterval = 20
)

// DefaultState returns a fully initialized game state.
// Engine code assumes this is the canonical zero-state.
func DefaultState() State {
	return State{
		Player: Player{
			Name:  "Traveller",
			Class: "Adventurer",
			Gold:  50,
			HP:    DefaultMaxHP,
			MaxHP: DefaultMaxHP,
			SP:    10,
			Level: 1,
			XP:    0,
			Inventory: map[string]int{
				"torch":        1,
				"rusty_dagger": 1,
			},
		},
		Meta: Meta{
			Location:        "Starting Village",
			QuestsCompleted: 0,
			CommandCount:    0,
		},
	}
}

// ================================
// State Helpers (Pure)
// ================================

// IsAlive returns true if the player has HP remaining.
func (p *Player) IsAlive() bool {
	return p.HP > 0
}

// ClampHP ensures HP does not exceed MaxHP or fall below zero.
func (p *Player) ClampHP() {
	if p.HP < 0 {
		p.HP = 0
	}
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// EnsureInventory guarantees the inventory map exists.
func (p *Player) EnsureInventory() {
	if p.Inventory == nil {
		p.Inventory = make(map[string]int)
	}
}
