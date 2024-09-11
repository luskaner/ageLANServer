package exec

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"golang.org/x/sys/windows"
	"path/filepath"
	"strings"
	"unsafe"
)

var (
	modshell32         = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteEx = modshell32.NewProc("ShellExecuteExW")
)

type SHELLEXECUTEINFO struct {
	cbSize         uint32
	fMask          uint32
	hwnd           windows.Handle
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       windows.Handle
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      windows.Handle
	dwHotKey       uint32
	hIconOrMonitor windows.Handle
	hProcess       windows.Handle
}

func shellExecuteEx(verb string, start bool, executable string, executableWorkingPath bool, getPid bool, show int32, arg ...string) (err error, pid uint32, exitCode int) {
	pid = 0
	exitCode = common.ErrSuccess
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(executable)
	args, _ := windows.UTF16PtrFromString(strings.Join(arg, " "))

	info := &SHELLEXECUTEINFO{
		cbSize:       uint32(unsafe.Sizeof(SHELLEXECUTEINFO{})),
		fMask:        0x00000040, // SEE_MASK_NOCLOSEPROCESS
		hwnd:         0,
		lpVerb:       verbPtr,
		lpFile:       exe,
		lpParameters: args,
		nShow:        show,
	}

	if executableWorkingPath {
		info.lpDirectory, _ = windows.UTF16PtrFromString(filepath.Dir(executable))
	}

	var ret uintptr
	ret, _, err = procShellExecuteEx.Call(uintptr(unsafe.Pointer(info)))
	if ret == 0 {
		return
	} else {
		err = nil
	}

	if !start {
		_, err = windows.WaitForSingleObject(info.hProcess, windows.INFINITE)
		if err != nil {
			return
		}
		var tmpExitCode uint32
		err = windows.GetExitCodeProcess(info.hProcess, &tmpExitCode)
		if err != nil {
			return
		}
		exitCode = int(tmpExitCode)
	} else if getPid {
		pid, err = windows.GetProcessId(info.hProcess)
	}

	return
}
