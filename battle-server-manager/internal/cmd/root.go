package cmd

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/pidLock"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "battle-server-manager manage Battle Server instances",
	Long:  "battle-server-manager manage Battle Server instances directly or as required by 'launcher'",
}

var Version string

func Execute() error {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		commonLogger.Println("Failed to lock pid file. Kill process 'battle-server-manager' if it is running in your task manager.")
		commonLogger.Println(err.Error())
		os.Exit(common.ErrPidLock)
	}
	defer func() {
		_ = lock.Unlock()
	}()
	RootCmd.Version = Version
	InitClean()
	InitRemove()
	InitRemoveAll()
	InitStart()
	return RootCmd.Execute()
}
