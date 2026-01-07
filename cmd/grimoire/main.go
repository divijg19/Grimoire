package main

import (
	"fmt"

	"github.com/divijg19/Grimoire/internal/adapters"
	"github.com/divijg19/Grimoire/internal/ui/cli"
)

func main() {
	store := adapters.NewJSONStore("grimoire.json")
	rng := adapters.NewMathRNG()

	state, err := store.Load()
	if err != nil {
		fmt.Println("Warning: load issue, continuing with defaults")
	}

	app := cli.NewApp(state, store, rng)
	app.Run()
}
