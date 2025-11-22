package gameLogs

import (
	"os"
	"path/filepath"
)

type GameAoE3 struct{}

func (g GameAoE3) Paths(path string) (paths []string) {
	logsPath := filepath.Join(path, "Logs")
	if f, err := os.Stat(logsPath); err != nil || !f.IsDir() {
		return
	}
	sessionDataFile := filepath.Join(logsPath, "Age3SessionData.txt")
	if f, err := os.Stat(sessionDataFile); err == nil && !f.IsDir() {
		paths = append(paths, sessionDataFile)
	}
	logFile := filepath.Join(logsPath, "Age3Log.txt")
	if f, err := os.Stat(logFile); err == nil && !f.IsDir() {
		paths = append(paths, logFile)
	}
	return
}
