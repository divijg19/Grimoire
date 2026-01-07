package ports

import "github.com/divijg19/Grimoire/internal/engine"

// Store abstracts game state persistence.
type Store interface {
	// Load returns a previously saved state.
	// If no save exists, it should return engine.DefaultState().
	Load() (*engine.State, error)

	// Save persists the given state atomically.
	Save(state *engine.State) error
}
