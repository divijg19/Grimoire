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
		// safety: ensure range is non-negative to avoid panic in RNG.Intn
		pRange := pMax - pMin + 1
		if pRange <= 0 {
			pRange = 1
		}
		pDmg := pMin + rng.Intn(pRange)

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
			// Victory: persist player's remaining HP into the state
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

			player.HP = playerHP
			return result, events
		}

		// ----------------
		// Enemy attack
		// ----------------
		eMin := enemy.AttackMin
		eMax := enemy.AttackMax
		// guard against malformed templates where max < min
		if eMax < eMin {
			eMax = eMin
		}
		eRange := eMax - eMin + 1
		if eRange <= 0 {
			eRange = 1
		}
		eDmg := eMin + rng.Intn(eRange)

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
