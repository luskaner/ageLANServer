package main

import (
	"os"
	"path/filepath"
	"scripts/internal"
	"scripts/internal/constants"

	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/game/cert"
)

func main() {
	for g := range game.SupportedGames.Iter() {
		gamePath := filepath.Join(constants.BuildDir, "mock", g)
		internal.MkdirP(gamePath)
		if ok, path := battleServer.ResolvePath(g); ok {
			internal.MkdirP(filepath.Join(gamePath, filepath.Dir(path)))
		}
		if ok, caCert := cert.NewCA(g, gamePath); ok {
			originalPath := caCert.OriginalPath()
			internal.MkdirP(filepath.Dir(originalPath))
			if _, err := os.Stat(originalPath); err != nil {
				if f, err := os.Create(originalPath); err != nil {
					panic(err)
				} else {
					_ = f.Close()
				}
			}
		}
	}
}
