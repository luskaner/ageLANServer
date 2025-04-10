package internal

import (
	"math/rand/v2"
	"sync"
	"time"
)

var seed = uint64(time.Now().UnixNano())
var src = rand.NewPCG(seed, seed)
var rng = rand.New(src)
var rngLock = sync.Mutex{}

func WithRng(action func(rand *rand.Rand)) {
	rngLock.Lock()
	defer rngLock.Unlock()
	action(rng)
}
