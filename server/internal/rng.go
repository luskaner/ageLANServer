package internal

import (
	"math/rand/v2"
	"sync"
	"time"
)

var seed = uint64(time.Now().UnixNano())
var src = rand.NewPCG(seed, seed)
var Rng = rand.New(src)
var RngLock = sync.Mutex{}
