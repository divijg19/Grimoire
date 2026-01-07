package cli

import (
	"fmt"

	"github.com/divijg19/Grimoire/internal/engine"
)

func (a *App) handle(events engine.Events, err error) {
	if err != nil {
		fmt.Println(cs("Error: "+err.Error(), bold, red))
		return
	}

	for _, e := range events {
		renderEvent(e)
	}

	// After handling events, show compact HP-only UI for minimal output.
	RenderHP(a.state)
	_ = a.store.Save(a.state)
}

func renderEvent(e engine.Event) {
	switch ev := e.(type) {

	case engine.EncounterStarted:
		fmt.Println(cs(fmt.Sprintf("You encountered a %s.", ev.EnemyID), bold, yellow))

	case engine.DamageDealt:
		if ev.Target == "player" {
			fmt.Println(c(fmt.Sprintf("You took %d damage (%d HP left).", ev.Amount, ev.HPLeft), red))
		} else {
			fmt.Println(c(fmt.Sprintf("You dealt %d damage (%d HP left).", ev.Amount, ev.HPLeft), green))
		}

	case engine.EnemyDefeated:
		fmt.Println(c(fmt.Sprintf("Enemy defeated! +%d XP, +%d gold.", ev.XP, ev.Gold), green))

	case engine.PlayerDefeated:
		fmt.Println(cs("You were defeated.", bold, red))

	case engine.LevelUp:
		fmt.Println(cs(fmt.Sprintf("Level up! Level %d. Max HP %d.", ev.NewLevel, ev.NewMaxHP), bold, magenta))

	case engine.ItemAdded:
		fmt.Println(c(fmt.Sprintf("Obtained %s x%d.", ev.ItemID, ev.Count), cyan))

	case engine.GoldGained:
		fmt.Println(c(fmt.Sprintf("Gained %d gold.", ev.Amount), yellow))
	}
}
