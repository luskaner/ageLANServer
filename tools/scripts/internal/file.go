package internal

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func Cp(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)
	if err := MkdirP(filepath.Dir(dst)); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, in)
	return err
}

func MkdirP(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}
	return nil
}
