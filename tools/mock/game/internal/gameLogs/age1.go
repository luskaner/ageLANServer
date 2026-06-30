package gameLogs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GameAoE1 struct{}

func (g GameAoE1) CreateLogs(path string) (err error) {
	logsPath := filepath.Join(path, "Logs")
	if err = os.MkdirAll(logsPath, os.ModePerm); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(logsPath, "StartupLog.txt"), []byte("Startup log content"), os.ModePerm); err != nil {
		return
	}
	return os.WriteFile(
		filepath.Join(logsPath, fmt.Sprintf("%s_base_log.txt", time.Now().Format(`2006.01.02-15.04.05`))),
		[]byte("Base log content"),
		os.ModePerm,
	)
}
