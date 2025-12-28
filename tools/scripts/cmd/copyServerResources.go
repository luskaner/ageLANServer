package main

import (
	"log"
	"os"
	"path/filepath"
	i "scripts/internal"

	e "github.com/luskaner/ageLANServer/common/executables"
)

func main() {
	module := e.Server
	src := i.ResourcePath(module)
	dst := i.BuildResourcePath(module)

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			i.MkdirP(target)
			return nil
		}
		return i.Cp(path, target)
	})
	if err != nil {
		log.Fatal(err)
	}
}
