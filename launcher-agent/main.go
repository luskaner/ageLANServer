package main

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	"github.com/luskaner/aoe2DELanServer/launcher-agent/internal"
	launcherCommonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"golang.org/x/sys/windows"
	"os"
	"os/signal"
	"strconv"
)

const revertCmdStart = 6

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	steamProcess, _ := strconv.ParseBool(os.Args[1])
	microsoftStoreProcess, _ := strconv.ParseBool(os.Args[2])
	serverExe := os.Args[3]
	broadcastBattleServer, _ := strconv.ParseBool(os.Args[4])
	revertCmdLength, _ := strconv.ParseInt(os.Args[5], 10, 64)
	revertCmdEnd := revertCmdStart + revertCmdLength
	var revertCmd []string
	if revertCmdLength > 0 {
		revertCmd = os.Args[revertCmdStart:revertCmdEnd]
	}
	var revertFlags []string
	if int64(len(os.Args)) > revertCmdEnd {
		revertFlags = os.Args[revertCmdEnd:]
	}
	var exitCode int
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
			if len(revertFlags) > 0 {
				internal.RunConfig(revertFlags)
			}
			if len(revertCmd) > 0 {
				launcherCommonExecutor.RunCommand(revertCmd)
			}
			_ = lock.Unlock()
			os.Exit(exitCode)
		}
	}()
	internal.Watch(steamProcess, microsoftStoreProcess, serverExe, broadcastBattleServer, revertFlags, revertCmd, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
