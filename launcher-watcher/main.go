package main

import (
	"common"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"os/signal"
	"strconv"
	"watcher/internal"
)

func main() {
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
			os.Exit(exitCode)
		}
	}()
	watchedProcess := os.Args[1]
	serverPid, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		log.Println("Failed to parse server pid.")
		exitCode = internal.ErrParseServerPid
		return
	}
	internal.Watch(watchedProcess, int(serverPid), revertFlags, &exitCode)
	os.Exit(exitCode)
}
