package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

func runClean(args []string) (err error, exitCode int) {
	fs := pflag.NewFlagSet("clean", pflag.ContinueOnError)
	cmd.GamesVarCommand(fs, &cmdUtils.GameIds)
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	commonLogger.Println("Cleaning up...")
	var games mapset.Set[string]
	games, err = cmdUtils.ParsedGameIds(nil)
	if err != nil {
		commonLogger.Println(err.Error())
		exitCode = internal.ErrGames
		return
	}
	var configs []battleServer.Config
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		configs, err = battleServer.Configs(g, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		if !cmdUtils.Remove(g, configs, true) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return
}
