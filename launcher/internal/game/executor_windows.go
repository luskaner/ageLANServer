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

func appNameSuffix(id string) string {
	switch id {
	case common.GameAoE1:
		return "Darwin"
	case common.GameAoE2:
		return "MSPhoenix"
	case common.GameAoE3:
		return "MSGPBoston"
	default:
		return ""
	}
}

func appName(id string) string {
	return appNamePrefix + appNameSuffix(id)
}

func appFamilyName(id string) string {
	return appName(id) + "_" + appPublisherId
}

func isInstalledOnXbox(id string) bool {
	packageFamilyName := appFamilyName(id)
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
		File:        fmt.Sprintf(`shell:appsfolder\%s!App`, appFamilyName(exec.gameId)),
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
