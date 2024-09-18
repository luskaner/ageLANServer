package internal

import (
	"os"
	"strings"
)

type CustomWriter struct {
	OriginalWriter *os.File
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "TLS handshake error") {
		return len(p), nil
	}
	return cw.OriginalWriter.Write(p)
}
