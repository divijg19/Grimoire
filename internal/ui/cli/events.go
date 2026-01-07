package cli

import (
	"fmt"

	"github.com/divijg19/Grimoire/internal/engine"
)

func (a *App) handle(events engine.Events, err error) {
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, e := range events {
		renderEvent(e)
	}

	RenderHUD(a.state)
	_ = a.store.Save(a.state)
}

func renderEvent(e engine.Event) {
	switch ev := e.(type) {

	case engine.EncounterStarted:
		fmt.Printf("You encountered a %s.\n", ev.EnemyID)

	case engine.DamageDealt:
		if ev.Target == "player" {
			fmt.Printf("You took %d damage (%d HP left).\n", ev.Amount, ev.HPLeft)
		} else {
			fmt.Printf("You dealt %d damage (%d HP left).\n", ev.Amount, ev.HPLeft)
		}

	case engine.EnemyDefeated:
		fmt.Printf("Enemy defeated! +%d XP, +%d gold.\n", ev.XP, ev.Gold)

	case engine.PlayerDefeated:
		fmt.Println("You were defeated.")

	case engine.LevelUp:
		fmt.Printf("Level up! Level %d. Max HP %d.\n", ev.NewLevel, ev.NewMaxHP)

	case engine.ItemAdded:
		fmt.Printf("Obtained %s x%d.\n", ev.ItemID, ev.Count)

	case engine.GoldGained:
		fmt.Printf("Gained %d gold.\n", ev.Amount)
	}
}
