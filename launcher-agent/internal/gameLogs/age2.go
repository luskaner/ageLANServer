package gameLogs

import (
	"os"
	"path/filepath"
)

type GameAoE2 struct{}

func (g GameAoE2) Paths(path string) (paths []string) {
	logsPath := filepath.Join(path, "logs")
	if f, err := os.Stat(logsPath); err != nil || !f.IsDir() {
		return
	}
	sessionDataFile := filepath.Join(logsPath, "Age2SessionData.txt")
	if f, err := os.Stat(sessionDataFile); err == nil && !f.IsDir() {
		paths = append(paths, sessionDataFile)
	}
	if matches, err := filepath.Glob(filepath.Join(logsPath, dateTimeNoDotGlob)); err == nil {
		addNewestPath(
			logsPath,
			matches,
			func(info os.FileInfo) bool {
				return info.IsDir()
			},
			&paths,
		)
	}
	return
}
