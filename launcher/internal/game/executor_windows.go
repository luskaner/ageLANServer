package game

import (
	"fmt"

	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/appx"
)

func (exec CustomExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	xboxProcess = true
	return
}

func (exec XboxExecutor) Execute(_ []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{
		File:        fmt.Sprintf(`shell:appsfolder\%s!App`, appx.FamilyName(exec.gameId)),
		Shell:       true,
		SpecialFile: true,
	}
	optionsFn(options)
	result = options.Exec()
	return
}

func (exec XboxExecutor) GamePath() string {
	return exec.gamePath
}

func (exec XboxExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	xboxProcess = true
	return
}

func startUri(uri string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: uri, Shell: true, SpecialFile: true}
	optionsFn(options)
	result = options.Exec()
	return
}
