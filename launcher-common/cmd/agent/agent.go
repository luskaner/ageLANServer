package agent

import (
	"runtime"

	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

type Values struct {
	*cmd.GameIdValues
	*cmd.LogRootValues
	ServerExecutable              string
	NoSteamProcess                bool
	XboxProcess                   bool
	BattleServerLANRebroadcast    bool
	BattleServerManagerExecutable string
	BattleServerRegion            string
	BaseDataPath                  string
}

func SingleFlagSet(version string, runFn func(*pflag.FlagSet) (err error, exitCode int)) (values *Values, singleFs *cmd.SingleFlagSet) {
	singleFs = cmd.NewSingleFlagSet(runFn, version)
	values = &Values{
		GameIdValues:  &cmd.GameIdValues{},
		LogRootValues: &cmd.LogRootValues{},
	}
	flags := singleFs.Fs()
	cmd.LogRootCommand(flags, values.LogRootRef())
	cmd.GameVarCommand(flags, values.GameIdRef())
	flags.StringVar(&values.ServerExecutable, "serverExecutable", "", "Path to the Server executable to for stopping it after game closes.")
	flags.StringVar(&values.BattleServerManagerExecutable, "bsManagerExecutable", "", "Path to the Battle Server Manager executable to monitor.")
	flags.StringVar(&values.BattleServerRegion, "bsRegion", "", "Region of the battle server for stopping it after game closes.")
	if runtime.GOOS == "windows" {
		flags.BoolVar(&values.NoSteamProcess, "noSteamProcess", false, "Whether to not monitor the Steam process for game start/stop events.")
		flags.BoolVar(&values.BattleServerLANRebroadcast, "bsLanRebroadcast", false, "Whether to rebroadcast LAN packets from the Battle Server to the local network.")
		flags.BoolVar(&values.XboxProcess, "xboxProcess", false, "Whether to monitor the Xbox process for game start/stop events.")
	}
	flags.StringVar(&values.BaseDataPath, "baseDataPath", "", "Base path for user data.")
	return
}
