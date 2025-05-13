package main

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcher_common "github.com/luskaner/ageLANServer/launcher-common"
	launcherCommonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

const revertCmdStart = 7

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	common.ChdirToExe()
	var steamProcess bool
	if runtime.GOOS == "windows" {
		steamProcess, _ = strconv.ParseBool(os.Args[1])
	} else {
		steamProcess = true
	}
	var xboxProcess bool
	if runtime.GOOS == "windows" {
		xboxProcess, _ = strconv.ParseBool(os.Args[2])
	}
	serverExe := os.Args[3]
	var broadcastBattleServer bool
	if runtime.GOOS == "windows" {
		broadcastBattleServer, _ = strconv.ParseBool(os.Args[4])
	}
	gameId := os.Args[5]
	revertCmdLength, _ := strconv.ParseInt(os.Args[6], 10, 64)
	revertCmdEnd := revertCmdStart + revertCmdLength
	var revertCmd []string
	if revertCmdLength > 0 {
		revertCmd = os.Args[revertCmdStart:revertCmdEnd]
	}
	var exitCode int
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
			defer func() {
				_ = lock.Unlock()
				os.Exit(exitCode)
			}()
			launcher_common.ConfigRevert(gameId, true, nil)
			if len(revertCmd) > 0 {
				_ = launcherCommonExecutor.RunRevertCommand(revertCmd)
			}
			if serverExe != "-" {
				_, _ = commonProcess.Kill(serverExe)
			}
		}
	}()
	watch.Watch(gameId, steamProcess, xboxProcess, serverExe, broadcastBattleServer, revertCmd, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
