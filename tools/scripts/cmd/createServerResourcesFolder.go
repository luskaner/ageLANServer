package main

import (
	"log"
	i "scripts/internal"
)

func main() {
	dst := i.BuildResourcePath("server")

	i.MkdirP(dst)
}
