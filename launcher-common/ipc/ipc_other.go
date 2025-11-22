//go:build !windows

package ipc

import (
	"os"
	"path/filepath"
)

func Path() string {
	return filepath.Join(os.TempDir(), name)
}
