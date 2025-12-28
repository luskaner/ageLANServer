package main

import (
	"log"
	i "scripts/internal"
)

func main() {
	dst := i.BuildResourcePath("server")

	if err := i.MkdirP(dst); err != nil {
		log.Fatal(err)
	}
}
