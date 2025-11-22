package gameLogs

import (
	"fmt"
	"os"
	"path/filepath"
)

type GameAoE1 struct{}

func (g GameAoE1) Paths(path string) (paths []string) {
	logsPath := filepath.Join(path, "Logs")
	if f, err := os.Stat(logsPath); err != nil || !f.IsDir() {
		return
	}
	startupLogFile := filepath.Join(logsPath, "StartupLog.txt")
	if f, err := os.Stat(startupLogFile); err == nil && !f.IsDir() {
		paths = append(paths, startupLogFile)
	}
	if matches, err := filepath.Glob(filepath.Join(logsPath, fmt.Sprintf("%s_base_log.txt", dateTimeGlob))); err != nil {
		return
	} else {
		addNewestPath(
			logsPath,
			matches,
			func(info os.FileInfo) bool {
				return !info.IsDir()
			},
			&paths,
		)
	}
	return
}
