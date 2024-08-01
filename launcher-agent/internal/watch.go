package internal

import (
	"common"
	commonProcess "common/process"
	"golang.org/x/sys/windows"
)

func waitForProcess(name string) bool {
	processes := commonProcess.ProcessesEntryNames([]string{name})
	if processes == nil || len(processes) == 0 {
		return false
	}
	process := processes[name]

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, true, process.ProcessID)

	if err != nil {
		return false
	}

	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	_, err = windows.WaitForSingleObject(handle, windows.INFINITE)

	if err != nil {
		return false
	}

	return true
}

func Watch(watchedProcess string, serverExe string, broadcastBattleServer bool, revertArgs []string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if len(revertArgs) > 0 {
		defer func() {
			RunConfig(revertArgs)
		}()
	}
	if !commonProcess.WaitUntilAnyProcessExist([]string{watchedProcess}) {
		*exitCode = ErrGameTimeoutStart
		return
	}
	if broadcastBattleServer {
		mostPriority, restInterfaces := common.RetrieveBsInterfaceAddresses()
		if mostPriority != nil && len(restInterfaces) > 0 {
			if !commonProcess.WaitUntilAnyProcessExist([]string{"BattleServer.exe"}) {
				*exitCode = ErrBattleServerTimeOutStart
				return
			}
			go CloneAnnouncements(mostPriority, restInterfaces)
		}
	}
	if waitForProcess(watchedProcess) {
		if serverExe != "-" {
			if _, err := commonProcess.Kill(serverExe); err != nil {
				*exitCode = ErrFailedStopServer
			}
		}
	} else {
		*exitCode = ErrFailedWaitForProcess
	}
}
