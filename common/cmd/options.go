package cmd

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/pflag"
	"strings"
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

func GameCommand(flags *pflag.FlagSet) {
	flags.StringP(
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

func GamesCommand(flags *pflag.FlagSet) {
	flags.StringArrayP(
		names,
		shorthand,
		common.SupportedGames.ToSlice(),
		fmt.Sprintf(
			`%s %s %s`,
			descriptionMultipleStart,
			strings.Join(common.SupportedGames.ToSlice(), ", "),
			descriptionEnd,
		),
	)
}
