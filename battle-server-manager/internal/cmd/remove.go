package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"fmt"
	"os"
	"slices"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove will kill a given Battle Server instance and remove the config file",
	Run: func(cmd *cobra.Command, args []string) {
		games, err := cmdUtils.ParsedGameIds(nil)
		if err != nil {
			os.Exit(internal.ErrGames)
		}
		region := viper.GetString("Region")
		for gameId := range games.Iter() {
			fmt.Printf("Game: %s\n", gameId)
			fmt.Printf("\tRemoving '%s' region...\n", region)
			configs, err := battleServerConfig.Configs(gameId, false)
			if err != nil {
				fmt.Printf("\t%s\n", err)
				continue
			}
			configs = slices.DeleteFunc(configs, func(c battleServerConfig.Config) bool {
				return c.Region != region
			})
			if !cmdUtils.Remove(gameId, configs, false) {
				fmt.Println("\tNo configuration needs it.")
			}
		}
	},
}

func InitRemove() {
	RemoveCmd.Flags().StringP(
		"region",
		"r",
		"",
		"Region of the battle server",
	)
	cmd.GamesVarCommand(RemoveCmd.Flags(), &cmdUtils.GameIds)
	err := RemoveCmd.MarkFlagRequired("region")
	if err != nil {
		panic(err)
	}
	if err = viper.BindPFlag("Region", RemoveCmd.Flags().Lookup("region")); err != nil {
		panic(err)
	}
	RootCmd.AddCommand(RemoveCmd)
}
