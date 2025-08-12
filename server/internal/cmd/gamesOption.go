package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
	"strings"
)

const Names = cmd.Name + `s`
const descriptionMultipleStart = cmd.DescriptionStart + `s`

type GameTitleValues struct {
	value mapset.Set[common.GameTitle]
}

func (g *GameTitleValues) Set(val string) error {
	var gameTitle common.GameTitle
	var err error
	if gameTitle, err = common.CheckGameTitleValue(val); err != nil {
		return err
	}
	if g.value == nil {
		g.value = mapset.NewThreadUnsafeSet[common.GameTitle]()
	}
	g.value.Add(gameTitle)
	return nil
}

func (g *GameTitleValues) Append(val string) error {
	var gameTitle common.GameTitle
	var err error
	if gameTitle, err = common.CheckGameTitleValue(val); err != nil {
		return err
	}
	g.value.Add(gameTitle)
	return nil
}

func (g *GameTitleValues) Replace(val []string) error {
	out := mapset.NewThreadUnsafeSet[common.GameTitle]()
	for _, d := range val {
		var gameTitle common.GameTitle
		var err error
		if gameTitle, err = common.CheckGameTitleValue(d); err != nil {
			return err
		}
		out.Add(gameTitle)
	}
	g.value = out
	return nil
}

func (g *GameTitleValues) GetSlice() []string {
	out := make([]string, g.value.Cardinality())
	for i, d := range g.value.ToSlice() {
		out[i] = string(d)
	}
	return out
}

func (g *GameTitleValues) Type() string {
	return "gameTitleSet"
}

func (g *GameTitleValues) String() string {
	return g.value.String()
}

func GamesCommand(flags *pflag.FlagSet, gameTitles *GameTitleValues) {
	flags.VarP(
		gameTitles,
		Names,
		cmd.Shorthand,
		fmt.Sprintf(
			`%s %s %s`,
			descriptionMultipleStart,
			strings.Join(cmd.SupportedGameSliceString(), ", "),
			cmd.DescriptionEnd,
		),
	)
}
