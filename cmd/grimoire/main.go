package main

import (
	"flag"
	"fmt"

	"github.com/divijg19/Grimoire/internal/adapters"
	"github.com/divijg19/Grimoire/internal/ui/cli"
	"github.com/divijg19/Grimoire/internal/ui/tui"
)

func main() {
	useCLI := flag.Bool("cli", false, "run legacy line-based CLI instead of fullscreen TUI")
	flag.Parse()

	store := adapters.NewJSONStore("grimoire.json")
	rng := adapters.NewMathRNG()

	state, err := store.Load()
	if err != nil {
		fmt.Println("Warning: load issue, continuing with defaults")
	}

	if *useCLI {
		app := cli.NewApp(state, store, rng)
		app.Run()
		return
	}

	app := tui.NewApp(state, store, rng)
	if err := app.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
