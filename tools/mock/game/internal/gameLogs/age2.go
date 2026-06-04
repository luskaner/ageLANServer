package gameLogs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type GameAoE2 struct{}

func (g GameAoE2) CreateLogs(path string) (err error) {
	logsPath := filepath.Join(path, "logs")
	if err = os.MkdirAll(logsPath, os.ModePerm); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(logsPath, "Age2SessionData.txt"), []byte("Age2 Session Data content"), os.ModePerm); err != nil {
		return
	}
	return os.WriteFile(
		filepath.Join(
			logsPath,
			fmt.Sprintf("%s_base_log.txt", time.Now().Format(`2006.01.02-1504.0000`)),
		),
		[]byte("Base log content"),
		os.ModePerm,
	)
}
