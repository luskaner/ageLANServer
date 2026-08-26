package cmd

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/fileLock"
	"github.com/luskaner/ageLANServer/common/logger"
)

var Version string
var rootFlagSet *cmd.RootFlagSet

type pidLocker interface {
	Lock() error
	Unlock() error
}

var (
	newPidLock     = func() pidLocker { return &fileLock.PidLock{} }
	newRootFlagSet = cmd.NewRootFlagSet
)

func Execute() (err error, exitCode int) {
	lock := newPidLock()
	if err = lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'battle-server-manager' if it is running in your task manager.")
		commonLogger.Println(err.Error())
		exitCode = common.ErrPidLock
		return
	}
	defer func() {
		_ = lock.Unlock()
	}()
	rootFlagSet = newRootFlagSet()
	rootFlagSet.RegisterCommand("clean", runClean)
	rootFlagSet.RegisterCommand("remove", runRemove)
	rootFlagSet.RegisterCommand("remove-all", runRemoveAll)
	rootFlagSet.RegisterCommand("start", runStart)
	return rootFlagSet.Execute(Version)
}
