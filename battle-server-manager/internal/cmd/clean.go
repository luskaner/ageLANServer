package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"os"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

func runClean(args []string) error {
	fs := pflag.NewFlagSet("clean", pflag.ContinueOnError)
	cmd.GamesVarCommand(fs, &cmdUtils.GameIds)
	if err := fs.Parse(args); err != nil {
		return err
	}

	commonLogger.Println("Cleaning up...")
	games, err := cmdUtils.ParsedGameIds(nil)
	if err != nil {
		commonLogger.Println(err.Error())
		os.Exit(internal.ErrGames)
	}
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		configs, err := battleServerConfig.Configs(g, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		if !cmdUtils.Remove(g, configs, true) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return nil
}
