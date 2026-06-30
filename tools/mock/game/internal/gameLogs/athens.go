package gameLogs

import (
	"os"
	"path/filepath"
)

type GameAoM struct{}

func (g GameAoM) CreateLogs(path string) (err error) {
	logsPath := filepath.Join(path, "temp", "Logs")
	if err = os.MkdirAll(logsPath, os.ModePerm); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(logsPath, "mythsessiondata.txt"), []byte("Myth Session Data content"), os.ModePerm); err != nil {
		return
	}
	return os.WriteFile(filepath.Join(logsPath, "mythlog.txt"), []byte("Myth Log content"), os.ModePerm)
}
