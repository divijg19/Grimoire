package adapters

import (
	"math/rand"
	"time"

	"github.com/divijg19/Grimoire/internal/ports"
)

// MathRNG is a math/rand-backed RNG adapter.
type MathRNG struct {
	r *rand.Rand
}

// NewMathRNG creates a new RNG seeded with current time.
func NewMathRNG() ports.RNG {
	return &MathRNG{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewSeededMathRNG creates a deterministic RNG (useful for tests / replays).
func NewSeededMathRNG(seed int64) ports.RNG {
	return &MathRNG{
		r: rand.New(rand.NewSource(seed)),
	}
}

func (m *MathRNG) Intn(n int) int {
	return m.r.Intn(n)
}

func (m *MathRNG) Float64() float64 {
	return m.r.Float64()
}
