package cmd

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
	"strings"
)

func GameVarCommand(flags *pflag.FlagSet, gameTitle *common.GameTitle) {
	flags.VarP(
		gameTitle,
		cmd.Name,
		cmd.Shorthand,
		fmt.Sprintf(
			`%s %s %s`,
			cmd.DescriptionStart,
			strings.Join(cmd.SupportedGameSliceString(), ", "),
			cmd.DescriptionEnd,
		),
	)
}
