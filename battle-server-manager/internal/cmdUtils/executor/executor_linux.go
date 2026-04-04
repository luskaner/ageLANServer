package executor

import (
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func modifyOptions(options *exec.Options) {
	wineOptions := exec.Options{
		File:        "wine",
		Args:        []string{"--version"},
		SpecialFile: true,
		ExitCode:    true,
		Wait:        true,
	}
	if result := wineOptions.Exec(); result.Success() {
		options.Args = append([]string{options.File}, options.Args...)
		options.File = wineOptions.File
		options.SpecialFile = wineOptions.SpecialFile
	}
}
