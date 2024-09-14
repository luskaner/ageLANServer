//go:build !windows

package common

func getExeFileName(name string) string {
	return name
}
