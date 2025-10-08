package hosts

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
)

const LineEnding = hosts.WindowsLineEnding

func FlushDns() (result *exec.Result) {
	result = exec.Options{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}.Exec()
	return
}

func Path() string {
	return filepath.Join(os.Getenv("WINDIR"), "System32", "drivers", "etc", "hosts")
}
