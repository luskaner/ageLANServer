package cmd

import (
	"battle-server-manager/internal"
	"battle-server-manager/internal/cmdUtils"
	"os"
	"slices"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove will kill a given Battle Server instance and remove the config file",
	Run: func(cmd *cobra.Command, args []string) {
		games, err := cmdUtils.ParsedGameIds(nil)
		if err != nil {
			commonLogger.Println(err.Error())
			os.Exit(internal.ErrGames)
		}
		region := viper.GetString("Region")
		for gameId := range games.Iter() {
			commonLogger.Printf("Game: %s\n", gameId)
			commonLogger.Printf("\tRemoving '%s' region...\n", region)
			configs, err := battleServerConfig.Configs(gameId, false)
			if err != nil {
				commonLogger.Printf("\t%s\n", err)
				continue
			}
			configs = slices.DeleteFunc(configs, func(c battleServerConfig.Config) bool {
				return c.Region != region
			})
			if !cmdUtils.Remove(gameId, configs, false) {
				commonLogger.Println("\tNo configuration needs it.")
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
