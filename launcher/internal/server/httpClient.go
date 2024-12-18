package server

import (
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"os"
	"path/filepath"
	"runtime"
)

func HttpGet(url string, insecureSkipVerify bool) int {
	args := []string{"-f", "-s", "-4"}
	if insecureSkipVerify {
		args = append(args, "-k")
	}
	args = append(args, url)
	var file string
	if runtime.GOOS == "windows" {
		// Make sure we use the built-in curl on Windows so it uses the system certificate store
		file = filepath.Join(os.Getenv("WINDIR"), "System32")
	}
	file = filepath.Join(file, "curl")
	result := exec.Options{
		File:        file,
		Args:        args,
		SpecialFile: true,
		Wait:        true,
		ExitCode:    true,
	}.Exec()
	return result.ExitCode
}
