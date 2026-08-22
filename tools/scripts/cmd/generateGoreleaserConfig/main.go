package main

import "scripts/internal/goreleaser"

func main() {
	err := goreleaser.Generate()
	if err != nil {
		panic(err)
	}
}
