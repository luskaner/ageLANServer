package internal

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"sync"

	"github.com/google/uuid"
)

var rng RandReader

// Deterministic mirrors the server's --deterministic flag. When enabled the
// server intentionally trades unpredictability for reproducibility (used to
// replay recorded sessions), so security-sensitive token generators fall back
// to the seeded PRNG. In normal operation it is false and those generators use
// a cryptographically secure source instead.
var Deterministic bool

// SecureBytes returns n random bytes. Outside deterministic mode they come from
// crypto/rand and are suitable for secrets such as session tokens.
func SecureBytes(n int) []byte {
	b := make([]byte, n)
	if Deterministic {
		WithRng(func(r *RandReader) {
			for i := range b {
				b[i] = byte(r.UintN(256))
			}
		})
		return b
	}
	if _, err := cryptorand.Read(b); err != nil {
		panic(err)
	}
	return b
}

// SecureIntN returns a uniformly distributed value in [0, n). Outside
// deterministic mode it is backed by crypto/rand.
func SecureIntN(n int) int {
	if n <= 0 {
		panic("SecureIntN: n must be positive")
	}
	if Deterministic {
		var v int
		WithRng(func(r *RandReader) {
			v = r.IntN(n)
		})
		return v
	}
	bound := uint64(n)
	// Largest multiple of bound representable in a uint64; rejecting values at
	// or above it removes modulo bias.
	limit := (math.MaxUint64/bound)*bound - 1
	var buf [8]byte
	for {
		if _, err := cryptorand.Read(buf[:]); err != nil {
			panic(err)
		}
		v := binary.BigEndian.Uint64(buf[:])
		if v <= limit {
			return int(v % bound)
		}
	}
}

type RandReader struct {
	*rand.Rand
	mu sync.Mutex
}

func (rr *RandReader) Read(p []byte) (int, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	for i := range p {
		p[i] = byte(rr.UintN(256))
	}
	return len(p), nil
}

func WithRng(action func(rand *RandReader)) {
	rng.WithRng(action)
}

func (rr *RandReader) WithRng(action func(rand *RandReader)) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	action(&rng)
}

func InitializeRng(seed uint64) {
	rng.Rand = rand.New(rand.NewPCG(seed, seed))
	uuid.SetRand(&rng)
}
