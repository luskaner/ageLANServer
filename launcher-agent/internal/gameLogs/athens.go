package gameLogs

import (
	"os"
	"path/filepath"
)

type GameAoM struct{}

func (g GameAoM) Paths(path string) (paths []string) {
	logsPath := filepath.Join(path, "temp", "Logs")
	if f, err := os.Stat(logsPath); err != nil || !f.IsDir() {
		return
	}
	sessionDataFile := filepath.Join(logsPath, "mythsessiondata.txt")
	if f, err := os.Stat(sessionDataFile); err == nil && !f.IsDir() {
		paths = append(paths, sessionDataFile)
	}
	logFile := filepath.Join(logsPath, "mythlog.txt")
	if f, err := os.Stat(logFile); err == nil && !f.IsDir() {
		paths = append(paths, logFile)
	}
	return
}
