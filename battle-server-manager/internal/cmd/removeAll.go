package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"os"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/cobra"
)

var RemoveAllCmd = &cobra.Command{
	Use:   "remove-all",
	Short: "remove-all will kill all Battle Server instances and config files",
	Run: func(cmd *cobra.Command, args []string) {
		commonLogger.Println("Removing all...")
		games, err := cmdUtils.ParsedGameIds(nil)
		if err != nil {
			commonLogger.Println(err.Error())
			os.Exit(internal.ErrGames)
		}
		for gameId := range games.Iter() {
			commonLogger.Printf("Game: %s\n", gameId)
			configs, err := battleServerConfig.Configs(gameId, false)
			if err != nil {
				commonLogger.Printf("\t%s\n", err)
				continue
			}
			if !cmdUtils.Remove(gameId, configs, false) {
				commonLogger.Println("\tNo configuration needs it.")
			}
		}
	},
}

func InitRemoveAll() {
	cmd.GamesVarCommand(RemoveAllCmd.Flags(), &cmdUtils.GameIds)
	RootCmd.AddCommand(RemoveAllCmd)
}
