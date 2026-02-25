package engine

import "errors"

// ================================
// Explore
// ================================

// Explore resolves a single explore action.
func Explore(state *State, rng RNG) (Events, error) {
	events := Events{}

	if !state.Player.IsAlive() {
		return events, errors.New("player is down (HP 0)")
	}

	state.Meta.CommandCount++

	roll := rng.Intn(100) + 1

	// Treasure (<=2%)
	if roll <= 2 {
		gold := 100 + rng.Intn(401) // 100–500
		state.Player.Gold += gold
		events = append(events,
			ExplorationResult{Kind: "treasure"},
			GoldGained{Amount: gold},
		)

		items := []string{"healing_potion", "rusty_dagger", "torch"}
		item := items[rng.Intn(len(items))]
		AddItem(state.PlayerPtr(), item, 1)
		events = append(events, ItemAdded{ItemID: item, Count: 1})

		return events, nil
	}

	// Item find (<=10%)
	if roll <= 10 {
		items := []string{"healing_potion", "torch"}
		item := items[rng.Intn(len(items))]
		AddItem(state.PlayerPtr(), item, 1)
		events = append(events,
			ExplorationResult{Kind: "item"},
			ItemAdded{ItemID: item, Count: 1},
		)
		return events, nil
	}

	// Gold find (<=30%)
	if roll <= 30 {
		gold := 5 + rng.Intn(46) // 5–50
		state.Player.Gold += gold
		events = append(events,
			ExplorationResult{Kind: "gold"},
			GoldGained{Amount: gold},
		)
		return events, nil
	}

	// Enemy encounter (<=50%)
	if roll <= 50 {
		enemyID := ChooseEnemy(state, 0, rng)
		enemy := Enemies[enemyID]

		result, combatEvents := ResolveCombat(state, enemy, rng)
		events = append(events, combatEvents...)

		state.Player.HP = max(state.Player.HP, 0)

		if result.Outcome == "win" {
			events = append(events, GrantXP(state, result.XP)...)

			state.Player.Gold += result.Gold
			events = append(events, GoldGained{Amount: result.Gold})

			for _, it := range result.Loot {
				AddItem(state.PlayerPtr(), it, 1)
				events = append(events, ItemAdded{ItemID: it, Count: 1})
			}
		}

		return events, nil
	}

	// Nothing
	events = append(events, ExplorationResult{Kind: "nothing"})
	return events, nil
}

// ================================
// Hunt
// ================================

// Hunt resolves a hunt action with optional extra SP stake.
func Hunt(state *State, extraSP int, rng RNG) (Events, error) {
	events := Events{}

	if !state.Player.IsAlive() {
		return events, errors.New("player is down (HP 0)")
	}

	if extraSP < 0 {
		extraSP = 0
	}
	if extraSP > HuntExtraSPMax {
		extraSP = HuntExtraSPMax
	}

	cost := HuntBaseSP + extraSP
	if state.Player.SP < cost {
		return events, errors.New("not enough SP")
	}

	state.Player.SP -= cost
	state.Meta.CommandCount++
	events = append(events, SPSpent{Amount: cost})

	enemyID := ChooseEnemy(state, extraSP, rng)
	enemy := Enemies[enemyID]

	result, combatEvents := ResolveCombat(state, enemy, rng)
	events = append(events, combatEvents...)

	if result.Outcome == "win" {
		mult := 1.0 + 0.25*float64(extraSP)
		xp := int(float64(result.XP) * mult)
		gold := int(float64(result.Gold) * mult)

		events = append(events, GrantXP(state, xp)...)

		state.Player.Gold += gold
		events = append(events, GoldGained{Amount: gold})

		for _, it := range result.Loot {
			AddItem(state.PlayerPtr(), it, 1)
			events = append(events, ItemAdded{ItemID: it, Count: 1})
		}

		// revive-to-1 rule
		if state.Player.HP == 0 {
			state.Player.HP = 1
			events = append(events, HPRestored{Amount: 1})
		}
	}

	return events, nil
}

// ================================
// Rest
// ================================

// Rest converts SP into HP.
func Rest(state *State, sp int) (Events, error) {
	events := Events{}

	if sp <= 0 {
		return events, errors.New("invalid SP amount")
	}
	if state.Player.SP < sp {
		return events, errors.New("not enough SP")
	}

	state.Player.SP -= sp
	events = append(events, SPSpent{Amount: sp})

	hpGain := sp * RestHPPerSP
	state.Player.HP += hpGain
	state.Player.ClampHP()

	events = append(events, HPRestored{Amount: hpGain})
	return events, nil
}

// ================================
// Use Item
// ================================

// UseItem applies item effects.
func UseItem(state *State, itemID string, rng RNG) (Events, error) {
	events := Events{}
	itemID = NormalizeItemID(itemID)
	if itemID == "" {
		return events, errors.New("invalid item id")
	}

	if !HasItem(&state.Player, itemID, 1) {
		return events, errors.New("item not in inventory")
	}

	item, ok := Items[itemID]
	if !ok {
		return events, errors.New("unknown item")
	}

	// Healing potion (only usable item for now)
	if itemID == "healing_potion" {
		hpGain := item.HPMin + rng.Intn(item.HPMax-item.HPMin+1)
		spGain := item.SPMin + rng.Intn(item.SPMax-item.SPMin+1)

		state.Player.HP += hpGain
		state.Player.SP += spGain
		state.Player.ClampHP()

		RemoveItem(&state.Player, itemID, 1)

		events = append(events,
			ItemRemoved{ItemID: itemID, Count: 1},
			HPRestored{Amount: hpGain},
		)

		return events, nil
	}

	return events, errors.New("item has no use effect")
}

// ================================
// Helpers
// ================================

// ChooseEnemy mirrors Python enemy selection logic.
func ChooseEnemy(state *State, extraSP int, rng RNG) string {
	pool := []string{"goblin", "skeleton", "bandit", "wolf", "bear", "orc"}
	weights := []int{25, 20, 15, 10, 5, 2}

	if state.Player.Level >= 3 {
		for i := range weights {
			weights[i] = max(5, weights[i]-5)
		}
		weights[2] += 5 // bandit bias
	}

	if extraSP > 0 {
		weights[0] = max(0, weights[0]-extraSP*8)
		weights[len(weights)-1] += extraSP * 8
	}

	total := 0
	for _, w := range weights {
		total += w
	}

	roll := rng.Intn(total)
	cum := 0
	for i, w := range weights {
		cum += w
		if roll < cum {
			return pool[i]
		}
	}

	return pool[0]
}

// PlayerPtr helper (clarity)
func (s *State) PlayerPtr() *Player {
	return &s.Player
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
