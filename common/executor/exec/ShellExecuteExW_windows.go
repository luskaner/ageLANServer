package exec

import (
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/sys/windows"
)

var (
	modshell32         = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteEx = modshell32.NewProc("ShellExecuteExW")
)

// Injectable Windows API functions for testing
var (
	waitForSingleObjectFn = windows.WaitForSingleObject
	getExitCodeProcessFn  = windows.GetExitCodeProcess
	getProcessIdFn        = windows.GetProcessId
	shellExecuteExCallFn  = procShellExecuteEx.Call
)

// SetWaitForSingleObjectFn sets a custom WaitForSingleObject for testing.
func SetWaitForSingleObjectFn(fn func(h windows.Handle, dwMilliseconds uint32) (uint32, error)) {
	waitForSingleObjectFn = fn
}

// SetGetExitCodeProcessFn sets a custom GetExitCodeProcess for testing.
func SetGetExitCodeProcessFn(fn func(h windows.Handle, exitCode *uint32) error) {
	getExitCodeProcessFn = fn
}

// SetGetProcessIdFn sets a custom GetProcessId for testing.
func SetGetProcessIdFn(fn func(h windows.Handle) (uint32, error)) {
	getProcessIdFn = fn
}

// SetShellExecuteExCallFn sets a custom ShellExecuteEx Call for testing.
func SetShellExecuteExCallFn(fn func(a ...uintptr) (uintptr, uintptr, error)) (restore func()) {
	orig := shellExecuteExCallFn
	if fn == nil {
		shellExecuteExCallFn = procShellExecuteEx.Call
	} else {
		shellExecuteExCallFn = fn
	}
	return func() { shellExecuteExCallFn = orig }
}

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
	args, _ := windows.UTF16PtrFromString(strings.Join(fixArgs(arg...), " "))

	// SHELLEXECUTEINFO is an in/out parameter: ShellExecuteExW writes hProcess,
	// hInstApp and friends back into it. It used to be a stack allocation whose
	// address was handed to the API as a uintptr, so if the goroutine's stack
	// grew while the call was in flight the OS wrote into the old, discarded
	// stack and this copy kept hProcess == 0. The symptom was
	// GetProcessId(0) failing with ERROR_INVALID_HANDLE, or a bogus exit code
	// on the Wait path, which is the runas path used to start the elevated
	// 'config-admin-agent'. Pinning forces the struct onto the heap (where the
	// collector does not move it) and keeps it, and the UTF-16 buffers its
	// fields reference, alive for the duration of the call.
	var info SHELLEXECUTEINFO
	var pinner runtime.Pinner
	pinner.Pin(&info)
	info.cbSize = uint32(unsafe.Sizeof(SHELLEXECUTEINFO{}))
	info.fMask = 0x00000040 // SEE_MASK_NOCLOSEPROCESS
	info.lpVerb = verbPtr
	info.lpFile = exe
	info.lpParameters = args
	info.nShow = show

	if executableWorkingPath {
		info.lpDirectory, _ = windows.UTF16PtrFromString(filepath.Dir(executable))
	}

	ret, _, err := shellExecuteExCallFn(uintptr(unsafe.Pointer(&info)), 0, 0)
	// Copy out what the API wrote before letting the pin go: once Unpin runs the
	// struct is free to be collected, and hProcess is still read below.
	hProcess := info.hProcess
	pinner.Unpin()
	runtime.KeepAlive(&info)
	if ret == 0 {
		return
	}

	err = nil

	if !start {
		_, err = waitForSingleObjectFn(hProcess, windows.INFINITE)
		if err != nil {
			return
		}
		var tmpExitCode uint32
		err = getExitCodeProcessFn(hProcess, &tmpExitCode)
		if err != nil {
			return
		}
		exitCode = int(tmpExitCode)
	} else if getPid {
		pid, err = getProcessIdFn(hProcess)
	}

	return
}
