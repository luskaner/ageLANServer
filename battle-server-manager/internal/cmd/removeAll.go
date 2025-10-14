package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/cobra"
)

var RemoveAllCmd = &cobra.Command{
	Use:   "remove-all",
	Short: "remove-all will kill all Battle Server instances and config files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Removing all...")
		games, err := cmdUtils.ParsedGameIds(nil)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(internal.ErrGames)
		}
		for gameId := range games.Iter() {
			fmt.Printf("Game: %s\n", gameId)
			configs, err := battleServerConfig.Configs(gameId, false)
			if err != nil {
				fmt.Printf("\t%s\n", err)
				continue
			}
			if !cmdUtils.Remove(gameId, configs, false) {
				fmt.Println("\tNo configuration needs it.")
			}
		}
	},
}

func InitRemoveAll() {
	cmd.GamesVarCommand(RemoveAllCmd.Flags(), &cmdUtils.GameIds)
	RootCmd.AddCommand(RemoveAllCmd)
}
