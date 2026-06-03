package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"os"
	"slices"

	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common/logger"
)

func runRemove(args []string) error {
	values, flags := bsManager.RemoveFlagSet()
	if err := flags.Parse(args); err != nil {
		return err
	}

	games, err := cmdUtils.ParsedGameIds(&values.GameIds)
	if err != nil {
		commonLogger.Println(err.Error())
		os.Exit(internal.ErrGames)
	}
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		commonLogger.Printf("\tRemoving '%s' region...\n", values.Region)
		configs, err := battleServer.Configs(g, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		configs = slices.DeleteFunc(configs, func(c battleServer.Config) bool {
			return c.Region != values.Region
		})
		if !cmdUtils.Remove(g, configs, false) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return nil
}
