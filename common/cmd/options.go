package cmd

import (
	"fmt"
	"strings"

	"github.com/luskaner/ageLANServer/common/game"
	"github.com/spf13/pflag"
)

const GameIdentifier = `game`
const GamesIdentifier = GameIdentifier + `s`
const shorthand = `e`
const descriptionStart = `Game type.`
const descriptionMultipleStart = `Game types.`
const descriptionEnd = `are supported.`

func GameVarCommand(flags *pflag.FlagSet, gameId *string) {
	flags.StringVarP(
		gameId,
		GameIdentifier,
		shorthand,
		"",
		fmt.Sprintf(
			`%s %s %s`,
			descriptionStart,
			strings.Join(game.SupportedGames.ToSlice(), ", "),
			descriptionEnd,
		),
	)
}

func gamesDescription() string {
	return fmt.Sprintf(
		`%s %s %s`,
		descriptionMultipleStart,
		strings.Join(game.SupportedGames.ToSlice(), ", "),
		descriptionEnd,
	)
}

func GamesVarCommand(flags *pflag.FlagSet, gameIds *[]string) {
	flags.StringArrayVarP(
		gameIds,
		GamesIdentifier,
		shorthand,
		game.SupportedGames.ToSlice(),
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

type LogRootBase interface {
	LogRootRef() *string
}

type LogRootValues struct {
	LogRoot string
}

func (v *LogRootValues) LogRootRef() *string {
	return &v.LogRoot
}

type GameIdBase interface {
	GameIdRef() *string
}

type GameIdValues struct {
	GameId string
}

func (v *GameIdValues) GameIdRef() *string {
	return &v.GameId
}
