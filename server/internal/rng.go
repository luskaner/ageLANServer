package internal

import (
	"math/rand/v2"
	"sync"
)

var rng *rand.Rand
var rngLock = sync.Mutex{}

func InitializeRng(seed uint64) {
	rng = rand.New(rand.NewPCG(seed, seed))
}

func WithRng(action func(rand *rand.Rand)) {
	rngLock.Lock()
	defer rngLock.Unlock()
	action(rng)
}
