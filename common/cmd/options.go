package cmd

import (
	"fmt"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/pflag"
)

const name = `game`
const names = name + `s`
const shorthand = `e`
const descriptionStart = `Game type.`
const descriptionMultipleStart = `Game types.`
const descriptionEnd = `are supported.`

func GameVarCommand(flags *pflag.FlagSet, gameId *string) {
	flags.StringVarP(
		gameId,
		name,
		shorthand,
		"",
		fmt.Sprintf(
			`%s %s %s`,
			descriptionStart,
			strings.Join(common.SupportedGames.ToSlice(), ", "),
			descriptionEnd,
		),
	)
}

func gamesDescription() string {
	return fmt.Sprintf(
		`%s %s %s`,
		descriptionMultipleStart,
		strings.Join(common.SupportedGames.ToSlice(), ", "),
		descriptionEnd,
	)
}

func GamesVarCommand(flags *pflag.FlagSet, gameIds *[]string) {
	flags.StringArrayVarP(
		gameIds,
		names,
		shorthand,
		common.SupportedGames.ToSlice(),
		gamesDescription(),
	)
}

func GamesCommand(flags *pflag.FlagSet) {
	flags.StringArrayP(
		names,
		shorthand,
		common.SupportedGames.ToSlice(),
		gamesDescription(),
	)
}

func LogRootCommand(flags *pflag.FlagSet, logRoot *string) {
	flags.StringVar(
		logRoot,
		"logRoot",
		"",
		"Path to the log folder. If not empty, enables extra logging.",
	)
}
