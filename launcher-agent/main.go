package main

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/pidLock"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/watch"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
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
	var rebroadcastIPs []net.IP
	if runtime.GOOS == "windows" {
		rebroadcastIPsStr := strings.Split(os.Args[4], ",")
		rebroadcastIPs = make([]net.IP, len(rebroadcastIPsStr))
		for i, ipStr := range rebroadcastIPsStr {
			if ip := net.ParseIP(ipStr); ip != nil {
				rebroadcastIPs[i] = ip
			}
		}
	}
	gameTitle := os.Args[5]
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
			launcherCommon.ConfigRevert(common.GameTitle(gameTitle), true, nil, nil)
			_ = launcherCommon.RunRevertCommand()
			if serverPid != 0 {
				_ = commonProcess.KillPid(serverPid)
			}
		}
	}()
	watch.Watch(common.GameTitle(gameTitle), steamProcess, xboxProcess, serverPid, rebroadcastIPs, &exitCode)
	_ = lock.Unlock()
	os.Exit(exitCode)
}
