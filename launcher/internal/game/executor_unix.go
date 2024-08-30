//go:build !windows

package game

import (
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"os"
	"path"
)

func isInstalledOnMicrosoftStore() bool {
	return false
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	steamProcess = true
	return
}

func (exec MicrosoftStoreExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	// Should not be called
	return
}

func (exec MicrosoftStoreExecutor) GameProcesses() (steamProcess bool, microsoftStoreProcess bool) {
	// Should not be called
	return
}

func steamInstallationPath() (path string) {
	path = homeDirPath()
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		path = ""
	}
	return
}

func homeDirPath() (p string) {
	if homeDir, err := os.UserHomeDir(); err == nil {
		p = path.Join(homeDir, dir)
	}
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: "open " + uri, Shell: true, SpecialFile: true}.Exec()
	return
}
