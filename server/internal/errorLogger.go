package internal

import (
	"io"
	"strings"
)

type CustomWriter struct {
	OriginalWriter io.Writer
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "TLS handshake error") {
		return len(p), nil
	}
	return cw.OriginalWriter.Write(p)
}
