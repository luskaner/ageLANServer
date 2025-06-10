package main

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

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
	var serverPid int
	if serverPidInt64, err := strconv.ParseInt(os.Args[3], 10, 32); err == nil {
		serverPid = int(serverPidInt64)
	}
	var broadcastBattleServer bool
	if runtime.GOOS == "windows" {
		broadcastBattleServer, _ = strconv.ParseBool(os.Args[4])
	}
	gameId := os.Args[5]
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
			launcherCommon.ConfigRevert(gameId, true, nil, nil)
			_ = launcherCommon.RunRevertCommand()
			if serverPid != 0 {
				_ = commonProcess.KillPid(serverPid)
			}
		}
	}()
	watch.Watch(gameId, steamProcess, xboxProcess, serverPid, broadcastBattleServer, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
