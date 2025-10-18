package internal

import (
	"io"
	"sync"
)

var PrefixerMutex sync.Mutex

type PrefixedWriter struct {
	writer io.Writer
	prefix []byte
	mutex  sync.Mutex
}

func NewPrefixedWriter(writer io.Writer, game string, name string) *PrefixedWriter {
	return &PrefixedWriter{
		writer: writer,
		prefix: []byte("[" + game + "][" + name + "] "),
	}
}

func (pw *PrefixedWriter) Write(p []byte) (n int, err error) {
	PrefixerMutex.Lock()
	defer PrefixerMutex.Unlock()
	_, err = pw.writer.Write(pw.prefix)
	if err != nil {
		return 0, err
	}

	n, err = pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}
