package cmd

import (
	"battle-server-manager/internal/cmdUtils"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

func runRemoveAll(args []string) (err error, exitCode int) {
	fs := pflag.NewFlagSet("remove-all", pflag.ContinueOnError)
	cmd.GamesVarCommand(fs, &cmdUtils.GameIds)
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	commonLogger.Println("Removing all...")
	return cmdUtils.RemoveGames(nil, false)
}
