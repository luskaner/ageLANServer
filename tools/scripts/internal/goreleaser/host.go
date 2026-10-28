package goreleaser

func windowsReleaseTarget(majorVersion uint32) OperatingSystem {
	if majorVersion < 10 {
		return OSWindowsLegacy
	}
	return OSWindowsModern
}
