package engine

// ================================
// RNG Port (used by combat)
// ================================

// RNG is injected to keep combat deterministic and testable.
type RNG interface {
	Intn(n int) int   // returns [0, n)
	Float64() float64 // returns [0.0, 1.0)
}

// ================================
// Combat Resolution
// ================================

// CombatResult summarizes terminal combat outcomes.
type CombatResult struct {
	Outcome string // "win" or "lose"
	XP      int
	Gold    int
	Loot    []string
}

// ResolveCombat runs a full combat loop between player and enemy template.
// - Player attacks first
// - Player damage scales with level
// - Enemy damage uses template ranges
// - Emits detailed combat events
func ResolveCombat(
	state *State,
	enemy EnemyTemplate,
	rng RNG,
) (CombatResult, Events) {

	events := Events{}
	player := &state.Player

	playerHP := player.HP
	enemyHP := enemy.HP
	level := player.Level

	// Encounter start
	events = append(events, EncounterStarted{EnemyID: enemy.ID})

	for playerHP > 0 && enemyHP > 0 {

		// ----------------
		// Player attack
		// ----------------
		pMin := 1 + level
		pMax := 2 + level
		pDmg := pMin + rng.Intn(pMax-pMin+1)

		enemyHP -= pDmg
		if enemyHP < 0 {
			enemyHP = 0
		}

		events = append(events, DamageDealt{
			Source: "player",
			Target: enemy.ID,
			Amount: pDmg,
			HPLeft: enemyHP,
		})

		if enemyHP <= 0 {
			// Victory
			result := CombatResult{
				Outcome: "win",
				XP:      enemy.XP,
				Gold:    enemy.Gold,
			}

			// Roll loot
			for _, drop := range enemy.Loot {
				if rng.Float64() < drop.Chance {
					result.Loot = append(result.Loot, drop.ItemID)
				}
			}

			events = append(events, EnemyDefeated{
				EnemyID: enemy.ID,
				XP:      enemy.XP,
				Gold:    enemy.Gold,
			})

			return result, events
		}

		// ----------------
		// Enemy attack
		// ----------------
		eMin := enemy.AttackMin
		eMax := enemy.AttackMax
		eDmg := eMin + rng.Intn(eMax-eMin+1)

		playerHP -= eDmg
		if playerHP < 0 {
			playerHP = 0
		}

		events = append(events, DamageDealt{
			Source: enemy.ID,
			Target: "player",
			Amount: eDmg,
			HPLeft: playerHP,
		})

		if playerHP <= 0 {
			// Defeat
			player.HP = 0
			events = append(events, PlayerDefeated{})
			return CombatResult{
				Outcome: "lose",
			}, events
		}
	}

	// Should not reach here, but be safe
	player.HP = playerHP
	return CombatResult{
		Outcome: "lose",
	}, events
}
