package cmd

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/common/logger"
)

var Version string
var rootFlagSet *cmd.RootFlagSet

func Execute() error {
	lock := &fileLock.PidLock{}
	if err := lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'battle-server-manager' if it is running in your task manager.")
		commonLogger.Println(err.Error())
		os.Exit(common.ErrPidLock)
	}
	defer func() {
		_ = lock.Unlock()
	}()

	rootFlagSet = cmd.NewRootFlagSet()
	rootFlagSet.RegisterCommand("clean", runClean)
	rootFlagSet.RegisterCommand("remove", runRemove)
	rootFlagSet.RegisterCommand("remove-all", runRemoveAll)
	rootFlagSet.RegisterCommand("start", runStart)
	return rootFlagSet.Execute(Version)
}
