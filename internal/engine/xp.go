package engine

// ================================
// XP & Leveling
// ================================

// XPToNext returns the XP required to reach the next level.
// Linear curve: level * 100
func XPToNext(level int) int {
	if level < 1 {
		level = 1
	}
	return level * 100
}

// GrantXP adds XP to the player, processes level-ups,
// mutates state, and emits progression events.
func GrantXP(state *State, amount int) Events {
	events := Events{}

	if amount <= 0 {
		return events
	}

	// Apply XP gain
	state.Player.XP += amount
	events = append(events, XPGained{Amount: amount})

	// Handle level-ups
	for {
		level := state.Player.Level
		need := XPToNext(level)

		if state.Player.XP < need {
			break
		}

		// Level up
		state.Player.XP -= need
		state.Player.Level++

		// Increase max HP
		state.Player.MaxHP += 10

		// Heal some HP on level-up (matches Python semantics)
		state.Player.HP += 10
		state.Player.ClampHP()

		events = append(events, LevelUp{
			NewLevel: state.Player.Level,
			NewMaxHP: state.Player.MaxHP,
		})
	}

	return events
}
