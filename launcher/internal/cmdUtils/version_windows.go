package cmdUtils

import (
	"fmt"
	"golang.org/x/sys/windows"
)

const minMajorVersion = 10
const minMinorVersion = 0
const minBuildNumber = 17763
const alias = "Windows 10 (1809 - Redstone 5)"

func CheckVersion() error {
	major, minor, build := windows.RtlGetNtVersionNumbers()
	if major > minMajorVersion {
		return nil
	}
	if major == minMajorVersion && minor > minMinorVersion {
		return nil
	}
	if major == minMajorVersion && minor == minMinorVersion && build >= minBuildNumber {
		return nil
	}
	return fmt.Errorf(
		"unsupported Windows version: %d.%d.%d, minimum is %d.%d.%d also known as %s. Update your system",
		major,
		minor,
		build,
		minMajorVersion,
		minMinorVersion,
		minBuildNumber,
		alias,
	)
}
