package appx

import (
	"os"
	"path/filepath"
	"unsafe"

	"github.com/luskaner/ageLANServer/common"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	appNamePrefix  = "Microsoft."
	appPublisherId = "8wekyb3d8bbwe"
)

var (
	modkernelbase                   = windows.NewLazyDLL("kernelbase.dll")
	procFindPackagesByPackageFamily = modkernelbase.NewProc("FindPackagesByPackageFamily")
)

func appxAPIAvailable() bool {
	return procFindPackagesByPackageFamily.Find() == nil
}

const (
	PackageFilterHead       uint32 = 0x10
	ErrorSuccess            uint32 = 0
	ErrorInsufficientBuffer uint32 = 122
)

func appNameSuffix(gameTitle string) string {
	switch gameTitle {
	case common.GameAoE1:
		return "Darwin"
	case common.GameAoE2:
		return "MSPhoenix"
	case common.GameAoE3:
		return "MSGPBoston"
	// FIXME: Add common.GameAoM
	default:
		return ""
	}
}

func name(gameTitle string) string {
	return appNamePrefix + appNameSuffix(gameTitle)
}

func FamilyName(gameTitle string) string {
	return name(gameTitle) + "_" + appPublisherId
}

func PackageFamilyNameToFullName(packageFamilyName string) (ok bool, fullName string) {
	if !appxAPIAvailable() {
		return
	}
	pfnUTF16, err := windows.UTF16PtrFromString(packageFamilyName)
	if err != nil {
		return
	}

	var count uint32
	var bufferLength uint32

	ret, _, _ := procFindPackagesByPackageFamily.Call(
		uintptr(unsafe.Pointer(pfnUTF16)),
		uintptr(PackageFilterHead),
		uintptr(unsafe.Pointer(&count)),
		0,
		uintptr(unsafe.Pointer(&bufferLength)),
		0,
		0,
		0,
	)

	result := uint32(ret)

	if result == ErrorInsufficientBuffer {
		fullNames := make([]*uint16, count)
		buffer := make([]uint16, bufferLength)

		ret, _, _ = procFindPackagesByPackageFamily.Call(
			uintptr(unsafe.Pointer(pfnUTF16)),
			uintptr(PackageFilterHead),
			uintptr(unsafe.Pointer(&count)),
			uintptr(unsafe.Pointer(&fullNames[0])),
			uintptr(unsafe.Pointer(&bufferLength)),
			uintptr(unsafe.Pointer(&buffer[0])),
			0,
			0,
		)

		if uint32(ret) == ErrorSuccess && count > 0 {
			fullName = windows.UTF16PtrToString(fullNames[0])
			ok = true
		}
	}
	return
}

func InstallLocation(packageFullName string) (ok bool, installLocation string) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`SOFTWARE\Classes\Local Settings\Software\Microsoft\Windows\CurrentVersion\AppModel\Repository\Packages\`+packageFullName,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return
	}
	defer func(key registry.Key) {
		_ = key.Close()
	}(key)
	installLocation, _, err = key.GetStringValue("PackageRootFolder")
	if err != nil {
		return
	}
	ok = true
	return
}

func GameInstallLocation(gameTitle string) (ok bool, gameLocation string) {
	var fullName string
	ok, fullName = PackageFamilyNameToFullName(FamilyName(gameTitle))
	if !ok {
		return
	}
	var installLocation string
	if ok, installLocation = InstallLocation(fullName); !ok {
		return
	} else {
		gameLocation = filepath.Join(installLocation, "Game")
		if f, err := os.Stat(gameLocation); err != nil || !f.IsDir() {
			return false, ""
		}
		return true, gameLocation
	}
}
