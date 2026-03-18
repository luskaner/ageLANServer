package hosts

import (
	"os"
	"path/filepath"
)

const LineEnding = WindowsLineEnding

func Path() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
}
