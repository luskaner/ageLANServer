package cmd

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/spf13/pflag"
	"strings"
)

const Name = `gameTitle`
const Names = Name + `s`
const shorthand = `e`
const descriptionStart = `GameTitle title.`
const descriptionMultipleStart = `GameTitle titles.`
const descriptionEnd = `are supported.`

func supportedGameSliceString() []string {
	slice := make([]string, len(common.SupportedGameTitleSlice))
	for i, v := range common.SupportedGameTitleSlice {
		slice[i] = string(v)
	}
	return slice
}

func GameVarCommand(flags *pflag.FlagSet, gameTitle *string) {
	flags.StringVarP(
		gameTitle,
		Name,
		shorthand,
		"",
		fmt.Sprintf(
			`%s %s %s`,
			descriptionStart,
			strings.Join(supportedGameSliceString(), ", "),
			descriptionEnd,
		),
	)
}

func GameCommand(flags *pflag.FlagSet) {
	flags.StringP(
		Name,
		shorthand,
		"",
		fmt.Sprintf(
			`%s %s %s`,
			descriptionStart,
			strings.Join(supportedGameSliceString(), ", "),
			descriptionEnd,
		),
	)
}

func GamesCommand(flags *pflag.FlagSet) {
	flags.StringArrayP(
		Names,
		shorthand,
		supportedGameSliceString(),
		fmt.Sprintf(
			`%s %s %s`,
			descriptionMultipleStart,
			strings.Join(supportedGameSliceString(), ", "),
			descriptionEnd,
		),
	)
}
