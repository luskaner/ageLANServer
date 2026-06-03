package executor

import (
	"io"

	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
)

func StartAgent(game string, steamProcess bool, xboxProcess bool, serverExe string, broadcastBattleServer bool,
	battleServerExe string, battleServerRegion string, basePath string, logRoot string, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	if logRoot == "" || basePath == "" {
		logRoot = ""
		basePath = ""
	}
	values, singleFs := agent.SingleFlagSet("", nil)
	values.BattleServerLANRebroadcast = broadcastBattleServer
	values.NoSteamProcess = !steamProcess
	values.XboxProcess = xboxProcess
	values.ServerExecutable = serverExe
	values.BattleServerManagerExecutable = battleServerExe
	values.BattleServerRegion = battleServerRegion
	values.BaseDataPath = basePath
	values.LogRoot = logRoot
	values.GameId = game
	options := exec.Options{
		File: executables.NativeFileName(false, executables.LauncherAgent),
		Pid:  true,
		Args: cmd.FlagSetToArgs(singleFs.Fs(), false),
	}
	optionsFn(options)
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}
