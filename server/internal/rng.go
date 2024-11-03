package internal

import (
	"math/rand"
	"sync"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())
var Rng = rand.New(src)
var RngLock = sync.Mutex{}
