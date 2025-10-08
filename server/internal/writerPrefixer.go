package internal

import (
	"io"
)

type PrefixedWriter struct {
	Writer io.Writer
	Prefix []byte
}

func (pw *PrefixedWriter) Write(p []byte) (n int, err error) {
	_, err = pw.Writer.Write(pw.Prefix)
	if err != nil {
		return 0, err
	}

	n, err = pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}
