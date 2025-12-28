package main

import (
	"io"
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
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(in *os.File) {
			_ = in.Close()
		}(in)
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer func(out *os.File) {
			_ = out.Close()
		}(out)
		_, err = io.Copy(out, in)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}
