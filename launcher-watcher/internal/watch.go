package internal

import (
	"common"
	"golang.org/x/sys/windows"
	"launcher-common/executor"
	"log"
)

func waitForProcess(name string) bool {
	processes := executor.ProcessesEntryNames([]string{name})
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

func Watch(watchedProcess string, serverPid uint32, revertArgs []string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if len(revertArgs) > 0 {
		defer func() {
			RunConfig(revertArgs)
		}()
	}
	if !executor.WaitUntilAnyProcessExist([]string{watchedProcess}) {
		log.Println("Failed to wait for process start or took longer than 1 minute.")
		*exitCode = ErrTimeoutStart
		return
	}
	if waitForProcess(watchedProcess) {
		if serverPid > 0 {
			if err := executor.Kill(int(serverPid)); err != nil {
				log.Println("Failed to stop server.")
				*exitCode = ErrFailedStopServer
			}
		}
	} else {
		log.Println("Failed to wait for process end.")
		*exitCode = ErrFailedWaitForProcess
	}
}
