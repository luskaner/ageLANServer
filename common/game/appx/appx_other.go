//go:build !windows

package appx

func GameInstallLocation(_ string) (ok bool, gameLocation string) {
	return
}
