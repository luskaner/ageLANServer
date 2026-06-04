package gameLogs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GameAoE4 struct{}

func (g GameAoE4) CreateLogs(path string) (err error) {
	if err = os.WriteFile(filepath.Join(path, "session_data.txt"), []byte("Session Data content"), os.ModePerm); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(path, "warnings.log"), []byte("Warnings content"), os.ModePerm); err != nil {
		return
	}
	logsPath := filepath.Join(path, "LogFiles")
	if err = os.MkdirAll(logsPath, os.ModePerm); err != nil {
		return
	}
	return os.WriteFile(
		filepath.Join(
			logsPath,
			fmt.Sprintf("unhandled.%s.txt", time.Now().Format(`2006-01-02.15-04-05`)),
		),
		[]byte("Unhandled content"),
		os.ModePerm,
	)
}
