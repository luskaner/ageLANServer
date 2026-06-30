package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common/logger"
)

func runRemove(args []string) (err error, exitCode int) {
	values, flags := bsManager.RemoveFlagSet()
	if err = flags.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	var games mapset.Set[string]
	games, err = cmdUtils.ParsedGameIds(&values.GameIds)
	if err != nil {
		commonLogger.Println(err.Error())
		exitCode = internal.ErrGames
		return
	}
	var configs []battleServer.Config
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		commonLogger.Printf("\tRemoving '%s' region...\n", values.Region)
		configs, err = battleServer.Configs(g, false)
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
	return
}
