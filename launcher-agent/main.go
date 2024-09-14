package main

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	"github.com/luskaner/aoe2DELanServer/launcher-agent/internal"
	"github.com/luskaner/aoe2DELanServer/launcher-agent/internal/watch"
	launcherCommonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
)

const revertCmdStart = 6

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	var steamProcess bool
	if runtime.GOOS == "windows" {
		steamProcess, _ = strconv.ParseBool(os.Args[1])
	} else {
		steamProcess = true
	}
	var microsoftStoreProcess bool
	if runtime.GOOS == "windows" {
		microsoftStoreProcess, _ = strconv.ParseBool(os.Args[2])
	}
	serverExe := os.Args[3]
	var broadcastBattleServer bool
	if runtime.GOOS == "windows" {
		broadcastBattleServer, _ = strconv.ParseBool(os.Args[4])
	}
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
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
			if len(revertFlags) > 0 {
				internal.RunConfig(revertFlags)
			}
			if len(revertCmd) > 0 {
				_ = launcherCommonExecutor.RunRevertCommand(revertCmd)
			}
			if serverExe != "-" {
				_, _ = commonProcess.Kill(serverExe)
			}
			_ = lock.Unlock()
			os.Exit(exitCode)
		}
	}()
	watch.Watch(steamProcess, microsoftStoreProcess, serverExe, broadcastBattleServer, revertFlags, revertCmd, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
