package main

import (
	"os"

	"scripts/internal/goreleaser"
)

func main() {
	currentOnly := len(os.Args) > 2 && os.Args[2] == "current"
	if len(os.Args) > 2 && !currentOnly {
		panic("unsupported target: " + os.Args[2])
	}
	generate := goreleaser.Generate
	if currentOnly {
		generate = goreleaser.GenerateCurrent
	}
	err := generate()
	if err != nil {
		panic(err)
	}
}
