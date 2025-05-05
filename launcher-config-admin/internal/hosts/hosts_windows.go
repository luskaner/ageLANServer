package hosts

import (
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"os"
	"path/filepath"
)

const LineEnding = hosts.WindowsLineEnding

func FlushDns() (result *exec.Result) {
	result = exec.Options{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}.Exec()
	return
}

func Path() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
}
