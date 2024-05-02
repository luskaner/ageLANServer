package internal

import (
	"math/rand"
	"time"
)

var src = rand.NewSource(time.Now().UnixNano())
var Rng = rand.New(src)
