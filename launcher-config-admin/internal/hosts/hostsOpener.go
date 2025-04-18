package hosts

import (
	"bufio"
	"fmt"
	"os"
)

type Opener struct {
	f       *os.File
	Mapping map[string]string
}

func (r *Opener) Open(path string, mode int) error {
	if r.f != nil {
		return fmt.Errorf("file already open")
	}
	f, err := os.OpenFile(
		path,
		mode,
		0666,
	)
	if err != nil {
		r.f = f
	}
	return err
}

func (r *Opener) Read() error {
	if r.f == nil {
		return fmt.Errorf("file not open")
	}
	var line string
	scanner := bufio.NewScanner(r.f)
	for scanner.Scan() {
		line = scanner.Text()
		lineIp, lineHost := mapping(line)
		if lineIp != "" && lineHost != "" {
			r.Mapping[lineHost] = lineIp
		}
	}
	return nil
}

func (r *Opener) Close() error {
	if r.f == nil {
		return nil
	}
	err := r.f.Close()
	if err != nil {
		r.f = nil
	}
	return err
}
