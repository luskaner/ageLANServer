package main

import (
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
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
	gameId := os.Args[5]
	serverExe := os.Args[3]
	var broadcastBattleServer bool
	if runtime.GOOS == "windows" && gameId != common.GameAoM {
		broadcastBattleServer, _ = strconv.ParseBool(os.Args[4])
	}
	battleServerExe := os.Args[6]
	battleServerRegion := os.Args[7]
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
			launcherCommon.ConfigRevert(gameId, true, nil)
			_ = launcherCommon.RunRevertCommand()
			if serverExe != "-" {
				_, _ = commonProcess.Kill(serverExe)
				if battleServerExe != "-" && battleServerRegion != "-" {
					launcherCommon.RemoveBattleServerRegion(battleServerExe, gameId, battleServerRegion)
				}
			}
		}
	}()
	watch.Watch(
		gameId,
		steamProcess,
		xboxProcess,
		serverExe,
		broadcastBattleServer,
		battleServerExe,
		battleServerRegion,
		&exitCode,
	)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
