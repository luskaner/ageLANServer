package main

import (
	"agent/internal"
	"common"
	"common/pidLock"
	"golang.org/x/sys/windows"
	"os"
	"os/signal"
)

func main() {
	lock := &pidLock.Lock{}
	if err := lock.Lock(); err != nil {
		os.Exit(common.ErrPidLock)
	}
	var exitCode int
	var revertFlags []string
	if len(os.Args) > 4 {
		revertFlags = os.Args[4:]
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
	internal.Watch(os.Args[1], os.Args[2], os.Args[3], revertFlags, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
