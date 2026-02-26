package engine

// ================================
// Item Catalog
// ================================

// Item represents a usable or collectible item.
// Behavior is defined elsewhere (rules), not here.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Optional use-effects (used by rules)
	HPMin int `json:"hp_min,omitempty"`
	HPMax int `json:"hp_max,omitempty"`
	SPMin int `json:"sp_min,omitempty"`
	SPMax int `json:"sp_max,omitempty"`
}

// Items is the global item registry.
var Items = map[string]Item{
	"healing_potion": {
		ID:    "healing_potion",
		Name:  "Healing Potion",
		HPMin: 10,
		HPMax: 25,
		SPMin: 1,
		SPMax: 3,
	},
	"torch": {
		ID:   "torch",
		Name: "Torch",
	},
	"rusty_dagger": {
		ID:   "rusty_dagger",
		Name: "Rusty Dagger",
	},
	"bone_shield": {
		ID:   "bone_shield",
		Name: "Bone Shield",
	},
	"ancient_coin": {
		ID:   "ancient_coin",
		Name: "Ancient Coin",
	},
	"coin_pouch": {
		ID:   "coin_pouch",
		Name: "Coin Pouch",
	},
	"wolf_pelt": {
		ID:   "wolf_pelt",
		Name: "Wolf Pelt",
	},
	"meat": {
		ID:    "meat",
		Name:  "Meat",
		HPMin: 40,
		HPMax: 40,
		SPMin: 2,
		SPMax: 2,
	},
	"bear_claw": {
		ID:   "bear_claw",
		Name: "Bear Claw",
	},
	"orcish_blade": {
		ID:   "orcish_blade",
		Name: "Orcish Blade",
	},
}

// ================================
// Enemy Catalog
// ================================

// LootEntry represents a probabilistic drop.
type LootEntry struct {
	ItemID string  `json:"item_id"`
	Chance float64 `json:"chance"` // 0.0–1.0
}

// EnemyTemplate defines a combat archetype.
type EnemyTemplate struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	HP        int         `json:"hp"`
	AttackMin int         `json:"attack_min"`
	AttackMax int         `json:"attack_max"`
	XP        int         `json:"xp"`
	Gold      int         `json:"gold"`
	Loot      []LootEntry `json:"loot"`
}

// Enemies is the global enemy registry.
var Enemies = map[string]EnemyTemplate{
	"goblin": {
		ID:        "goblin",
		Name:      "Goblin",
		HP:        8,
		AttackMin: 1,
		AttackMax: 3,
		XP:        5,
		Gold:      3,
		Loot: []LootEntry{
			{ItemID: "rusty_dagger", Chance: 0.20},
			{ItemID: "healing_potion", Chance: 0.10},
		},
	},
	"skeleton": {
		ID:        "skeleton",
		Name:      "Skeleton",
		HP:        10,
		AttackMin: 2,
		AttackMax: 4,
		XP:        8,
		Gold:      5,
		Loot: []LootEntry{
			{ItemID: "bone_shield", Chance: 0.10},
			{ItemID: "ancient_coin", Chance: 0.25},
		},
	},
	"bandit": {
		ID:        "bandit",
		Name:      "Bandit",
		HP:        12,
		AttackMin: 2,
		AttackMax: 5,
		XP:        10,
		Gold:      8,
		Loot: []LootEntry{
			{ItemID: "coin_pouch", Chance: 0.30},
			{ItemID: "healing_potion", Chance: 0.15},
		},
	},
	"wolf": {
		ID:        "wolf",
		Name:      "Wolf",
		HP:        14,
		AttackMin: 3,
		AttackMax: 6,
		XP:        12,
		Gold:      6,
		Loot: []LootEntry{
			{ItemID: "wolf_pelt", Chance: 0.30},
			{ItemID: "meat", Chance: 0.40},
		},
	},
	"bear": {
		ID:        "bear",
		Name:      "Bear",
		HP:        20,
		AttackMin: 4,
		AttackMax: 8,
		XP:        20,
		Gold:      10,
		Loot: []LootEntry{
			{ItemID: "bear_claw", Chance: 0.25},
		},
	},
	"orc": {
		ID:        "orc",
		Name:      "Orc",
		HP:        25,
		AttackMin: 5,
		AttackMax: 10,
		XP:        25,
		Gold:      15,
		Loot: []LootEntry{
			{ItemID: "orcish_blade", Chance: 0.15},
			{ItemID: "coin_pouch", Chance: 0.25},
		},
	},
}
