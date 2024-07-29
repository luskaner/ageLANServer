package main

import (
	"common"
	"common/pidLock"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/signal"
	"watcher/internal"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		fmt.Println("Failed to lock pid file. You may try checking if the process in PID file exists (and killing it).")
		os.Exit(common.ErrPidLock)
	}
	var exitCode int
	var revertFlags []string
	if len(os.Args) > 3 {
		revertFlags = os.Args[3:]
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
			internal.RunConfig(revertFlags)
			_ = lock.Unlock()
			os.Exit(exitCode)
		}
	}()
	watchedProcess := os.Args[1]
	internal.Watch(watchedProcess, os.Args[2], revertFlags, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
