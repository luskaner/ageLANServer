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
	if currentOnly {
		if err := goreleaser.GenerateCurrent(); err != nil {
			panic(err)
		}
		return
	}
	if err := goreleaser.Generate(); err != nil {
		panic(err)
	}
}
