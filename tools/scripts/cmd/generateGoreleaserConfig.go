package main

import (
	"os"

	"scripts/internal/goreleaser"
)

func main() {
	windowsOnly := len(os.Args) > 2 && os.Args[2] == "windows"
	if len(os.Args) > 2 && !windowsOnly {
		panic("unsupported target: " + os.Args[2])
	}
	err := goreleaser.Generate(windowsOnly)
	if err != nil {
		panic(err)
	}
}
