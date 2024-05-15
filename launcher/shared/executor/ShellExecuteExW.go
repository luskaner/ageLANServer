package executor

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
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

func ShellExecuteEx(lpExecInfo *SHELLEXECUTEINFO) error {
	ret, _, err := procShellExecuteEx.Call(uintptr(unsafe.Pointer(lpExecInfo)))
	if ret == 0 {
		return err
	}
	return nil
}

func ShellExecuteAndWait(verb string, executable string, arg ...string) bool {
	verbPtr, _ := windows.UTF16PtrFromString(verb)
	exe, _ := windows.UTF16PtrFromString(executable)
	args, _ := windows.UTF16PtrFromString(strings.Join(arg, " "))
	show := windows.SW_HIDE

	info := &SHELLEXECUTEINFO{
		cbSize:       uint32(unsafe.Sizeof(SHELLEXECUTEINFO{})),
		fMask:        0x00000040, // SEE_MASK_NOCLOSEPROCESS
		hwnd:         0,
		lpVerb:       verbPtr,
		lpFile:       exe,
		lpParameters: args,
		nShow:        int32(show),
	}

	err := ShellExecuteEx(info)
	if err != nil {
		return false
	}

	_, err = windows.WaitForSingleObject(info.hProcess, windows.INFINITE)
	if err != nil {
		return false
	}

	return true
}
