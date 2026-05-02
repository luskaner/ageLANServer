package executor

import (
	"io"
	"strconv"

	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func StartAgent(game string, steamProcess bool, xboxProcess bool, serverExe string, broadcastBattleServer bool,
	battleServerExe string, battleServerRegion string, basePath string, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	if serverExe == "" {
		serverExe = "-"
	}
	if battleServerExe == "" {
		battleServerExe = "-"
	}
	if battleServerRegion == "" {
		battleServerRegion = "-"
	}
	if logRoot == "" || basePath == "" {
		logRoot = "-"
		basePath = "-"
	}
	args := []string{
		strconv.FormatBool(steamProcess),
		strconv.FormatBool(xboxProcess),
		serverExe,
		strconv.FormatBool(broadcastBattleServer),
		game,
		battleServerExe,
		battleServerRegion,
		logRoot,
		basePath,
	}
	options := exec.Options{File: executables.NativeFileName(false, executables.LauncherAgent), Pid: true, Args: args}
	optionsFn(options)
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}
