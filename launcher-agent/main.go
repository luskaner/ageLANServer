package main

import (
	"agent/internal"
	"common"
	"common/pidLock"
	"golang.org/x/sys/windows"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	steamProcess, _ := strconv.ParseBool(os.Args[1])
	microsoftStoreProcess, _ := strconv.ParseBool(os.Args[2])
	serverExe := os.Args[3]
	broadcastBattleServer, _ := strconv.ParseBool(os.Args[4])
	var revertFlags []string
	if len(os.Args) > 5 {
		revertFlags = os.Args[5:]
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
			_ = lock.Unlock()
			os.Exit(exitCode)
		}
	}()
	internal.Watch(steamProcess, microsoftStoreProcess, serverExe, broadcastBattleServer, revertFlags, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
