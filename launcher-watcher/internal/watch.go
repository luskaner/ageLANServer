package internal

import (
	"common"
	commonProcess "common/process"
	"fmt"
	"golang.org/x/sys/windows"
	"log"
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

func Watch(watchedProcess string, serverExe string, revertArgs []string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if len(revertArgs) > 0 {
		defer func() {
			RunConfig(revertArgs)
		}()
	}
	if !commonProcess.WaitUntilAnyProcessExist([]string{watchedProcess}) {
		log.Println("Failed to wait for process start or took longer than 1 minute.")
		*exitCode = ErrTimeoutStart
		return
	}
	if waitForProcess(watchedProcess) {
		if len(serverExe) > 0 {
			if proc, err := commonProcess.Kill(serverExe); err != nil {
				log.Println("Failed to stop server.")
				if proc != nil {
					fmt.Println("You may try killing it manually. Search for the process with PID", proc.Pid)
				}
				*exitCode = ErrFailedStopServer
			}
		}
	} else {
		log.Println("Failed to wait for process end.")
		*exitCode = ErrFailedWaitForProcess
	}
}
