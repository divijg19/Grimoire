package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/divijg19/Grimoire/internal/engine"
	"github.com/divijg19/Grimoire/internal/ports"
)

type App struct {
	state *engine.State
	store ports.Store
	rng   ports.RNG
}

func NewApp(state *engine.State, store ports.Store, rng ports.RNG) *App {
	return &App{
		state: state,
		store: store,
		rng:   rng,
	}
}

func (a *App) Run() {
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println(cs("Grimoire — interactive mode. Type 'help'.", bold, cyan))
	RenderHUD(a.state)

	for {
		fmt.Print(cs("> ", bold, cyan))
		if !reader.Scan() {
			fmt.Println(cs("\nExiting and saving...", yellow))
			_ = a.store.Save(a.state)
			return
		}

		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		a.dispatch(line)
	}
}
