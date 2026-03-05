package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"os"
	"slices"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

var region string

func runRemove(args []string) error {
	fs := pflag.NewFlagSet("remove", pflag.ContinueOnError)
	fs.StringVarP(&region, "region", "r", "", "Region of the battle server")
	cmd.GamesVarCommand(fs, &cmdUtils.GameIds)
	if err := fs.Parse(args); err != nil {
		return err
	}

	games, err := cmdUtils.ParsedGameIds(nil)
	if err != nil {
		commonLogger.Println(err.Error())
		os.Exit(internal.ErrGames)
	}
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		commonLogger.Printf("\tRemoving '%s' region...\n", region)
		configs, err := battleServerConfig.Configs(g, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		configs = slices.DeleteFunc(configs, func(c battleServerConfig.Config) bool {
			return c.Region != region
		})
		if !cmdUtils.Remove(g, configs, false) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return nil
}
