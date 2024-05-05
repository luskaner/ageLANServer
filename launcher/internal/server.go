package internal

import "os"

func RunServer(config ServerConfig) bool {
	if !config.Start {
		return true
	}
	var exePath string
	if config.Executable == "" {
		dir, err := os.Getwd()
		if err != nil {
			return false
		}
		exePath = dir + `\server.exe`
	} else {
		exePath = config.Executable
	}

	if _, err := os.Stat(exePath); err != nil {
		return false
	}

	return RunCustomExecutable(exePath)
}
