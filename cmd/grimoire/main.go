package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/divijg19/Grimoire/internal/adapters"
	"github.com/divijg19/Grimoire/internal/engine"
	"github.com/divijg19/Grimoire/internal/ports"
)

// ================================
// Entry
// ================================

func main() {
	store := adapters.NewJSONStore("grimoire.json")
	rng := adapters.NewMathRNG()

	state, err := store.Load()
	if err != nil {
		fmt.Println("Warning: load issue, continuing with defaults")
	}

	repl(state, store, rng)
}

// ================================
// REPL
// ================================

func repl(state *engine.State, store ports.Store, rng engine.RNG) {
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println("Grimoire — interactive mode. Type 'help'.")

	for {
		fmt.Print("> ")
		if !reader.Scan() {
			fmt.Println("\nExiting and saving...")
			_ = store.Save(state)
			return
		}

		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]
		params := args[1:]

		switch cmd {

		case "help":
			printHelp()

		case "status":
			renderStatus(state)

		case "explore":
			events, err := engine.Explore(state, rng)
			handle(events, err, store, state)

		case "hunt":
			extra := 0
			if len(params) > 0 {
				fmt.Sscanf(params[0], "%d", &extra)
			}
			events, err := engine.Hunt(state, extra, rng)
			handle(events, err, store, state)

		case "rest":
			sp := 1
			if len(params) > 0 {
				fmt.Sscanf(params[0], "%d", &sp)
			}
			events, err := engine.Rest(state, sp)
			handle(events, err, store, state)

		case "use":
			if len(params) == 0 {
				fmt.Println("Usage: use <item_id>")
				continue
			}
			events, err := engine.UseItem(state, params[0], rng)
			handle(events, err, store, state)

		case "save":
			_ = store.Save(state)
			fmt.Println("Game saved.")

		case "exit", "quit":
			_ = store.Save(state)
			fmt.Println("Game saved. Goodbye.")
			return

		default:
			fmt.Println("Unknown command. Type 'help'.")
		}
	}
}

// ================================
// Event Handling
// ================================

func handle(events engine.Events, err error, store ports.Store, state *engine.State) {
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, e := range events {
		renderEvent(e)
	}

	_ = store.Save(state)
}

// ================================
// Rendering
// ================================

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

	case engine.XPGained:
		fmt.Printf("Gained %d XP.\n", ev.Amount)

	case engine.LevelUp:
		fmt.Printf("Level up! Level %d. Max HP %d.\n", ev.NewLevel, ev.NewMaxHP)

	case engine.ItemAdded:
		fmt.Printf("Obtained %s x%d.\n", ev.ItemID, ev.Count)

	case engine.ItemRemoved:
		fmt.Printf("Used %s.\n", ev.ItemID)

	case engine.GoldGained:
		fmt.Printf("Gained %d gold.\n", ev.Amount)

	case engine.HPRestored:
		fmt.Printf("Recovered %d HP.\n", ev.Amount)

	case engine.SPSpent:
		fmt.Printf("Spent %d SP.\n", ev.Amount)

	case engine.ExplorationResult:
		if ev.Kind == "nothing" {
			fmt.Println("You found nothing of interest.")
		}

	default:
		// silent ignore for future events
	}
}

// ================================
// Status / Help
// ================================

func renderStatus(state *engine.State) {
	p := state.Player
	fmt.Printf(
		"%s (%s) — Lv %d | HP %d/%d | SP %d | Gold %d | XP %d\n",
		p.Name, p.Class, p.Level, p.HP, p.MaxHP, p.SP, p.Gold, p.XP,
	)
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  status")
	fmt.Println("  explore")
	fmt.Println("  hunt [extra_sp]")
	fmt.Println("  rest [sp]")
	fmt.Println("  use <item_id>")
	fmt.Println("  save")
	fmt.Println("  exit / quit")
}
