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

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean removes all config files of non-running Battle Server instances",
	Long:  "clean removes all config files of non-running Battle Server instances (or any game or a specific one)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning up...")
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
			if !cmdUtils.Remove(gameId, configs, true) {
				fmt.Println("\tNo configuration needs it.")
			}
		}
	},
}

func InitClean() {
	cmd.GamesVarCommand(CleanCmd.Flags(), &cmdUtils.GameIds)
	RootCmd.AddCommand(CleanCmd)
}
