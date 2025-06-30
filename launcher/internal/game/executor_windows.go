package game

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"golang.org/x/sys/windows"
	"unsafe"
)

var (
	modkernelbase                   = windows.NewLazyDLL("kernelbase.dll")
	procFindPackagesByPackageFamily = modkernelbase.NewProc("FindPackagesByPackageFamily")
)

const (
	PackageFilterHead       uint32 = 0x10
	ErrorSuccess            uint32 = 0
	ErrorInsufficientBuffer uint32 = 122
)

const (
	appNamePrefix  = "Microsoft."
	appPublisherId = "8wekyb3d8bbwe"
)

func appNameSuffix(gameTitle common.GameTitle) string {
	switch gameTitle {
	case common.AoE1:
		return "Darwin"
	case common.AoE2:
		return "MSPhoenix"
	case common.AoE3:
		return "MSGPBoston"
	default:
		return ""
	}
}

func appName(gameTitle common.GameTitle) string {
	return appNamePrefix + appNameSuffix(gameTitle)
}

func appFamilyName(gameTitle common.GameTitle) string {
	return appName(gameTitle) + "_" + appPublisherId
}

func isInstalledOnXbox(gameTitle common.GameTitle) bool {
	packageFamilyName := appFamilyName(gameTitle)
	errLoadDll := modkernelbase.Load()
	if errLoadDll != nil {
		return false
	}
	if procFindPackagesByPackageFamily.Find() != nil {
		return false
	}
	pfnUTF16, err := windows.UTF16PtrFromString(packageFamilyName)
	if err != nil {
		return false
	}
	var count uint32
	var bufferLength uint32

	result, _, _ := procFindPackagesByPackageFamily.Call(
		uintptr(unsafe.Pointer(pfnUTF16)),
		uintptr(PackageFilterHead),
		uintptr(unsafe.Pointer(&count)),
		uintptr(0),
		uintptr(unsafe.Pointer(&bufferLength)),
		uintptr(0),
		uintptr(0),
	)

	apiReturnCode := uint32(result)
	if apiReturnCode == ErrorSuccess || apiReturnCode == ErrorInsufficientBuffer {
		return count > 0
	}
	return false
}

func (exec CustomExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	xboxProcess = true
	return
}

func (exec XboxExecutor) Execute(_ []string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{
		File:        fmt.Sprintf(`shell:appsfolder\%s!App`, appFamilyName(exec.gameTitle)),
		Shell:       true,
		SpecialFile: true,
	}.Exec()
	return
}

func (exec XboxExecutor) GameProcesses() (steamProcess bool, xboxProcess bool) {
	xboxProcess = true
	return
}

func startUri(uri string) (result *commonExecutor.Result) {
	result = commonExecutor.Options{File: uri, Shell: true, SpecialFile: true}.Exec()
	return
}
