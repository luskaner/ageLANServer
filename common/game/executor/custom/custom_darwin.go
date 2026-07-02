package custom

import "runtime"

func (exec Exec) GameProcesses() (steamProcess bool, steamMacOsNative bool, xboxProcess bool) {
	steamProcess = true
	// Only supported on Apple Silicon
	steamMacOsNative = runtime.GOARCH == "arm64"
	return
}
