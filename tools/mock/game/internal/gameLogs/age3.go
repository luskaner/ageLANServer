package gameLogs

import (
	"os"
	"path/filepath"
)

type GameAoE3 struct{}

func (g GameAoE3) CreateLogs(path string) (err error) {
	logsPath := filepath.Join(path, "Logs")
	if err = os.MkdirAll(logsPath, os.ModePerm); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(logsPath, "Age3SessionData.txt"), []byte("Age3 Session Data content"), os.ModePerm); err != nil {
		return
	}
	return os.WriteFile(filepath.Join(logsPath, "Age3Log.txt"), []byte("Age3 Log content"), os.ModePerm)
}
