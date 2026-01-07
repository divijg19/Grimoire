package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/divijg19/Grimoire/internal/engine"
)

func (a *App) dispatch(line string) {
	parts := strings.Fields(line)
	cmd := parts[0]
	args := parts[1:]

	var (
		events engine.Events
		err    error
	)

	switch cmd {

	case "help":
		PrintHelp()
		return

	case "status":
		RenderHUD(a.state)
		return

	case "explore":
		events, err = engine.Explore(a.state, a.rng)

	case "hunt":
		extra := 0
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &extra)
		}
		events, err = engine.Hunt(a.state, extra, a.rng)

	case "rest":
		sp := 1
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &sp)
		}
		events, err = engine.Rest(a.state, sp)

	case "use":
		if len(args) == 0 {
			fmt.Println(c("Usage: use <item_id>", yellow))
			return
		}
		events, err = engine.UseItem(a.state, args[0], a.rng)

	case "save":
		_ = a.store.Save(a.state)
		fmt.Println(c("Game saved.", green))
		return

	case "exit", "quit":
		_ = a.store.Save(a.state)
		fmt.Println(c("Game saved. Goodbye.", green))
		os.Exit(0)

	default:
		fmt.Println(c("Unknown command. Type 'help'.", yellow))
		return
	}

	a.handle(events, err)
}
