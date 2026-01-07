package ports

// RNG defines the randomness interface used by the engine.
// Implementations live in adapters/.
type RNG interface {
	// Intn returns an integer in [0, n).
	Intn(n int) int

	// Float64 returns a float in [0.0, 1.0).
	Float64() float64
}
