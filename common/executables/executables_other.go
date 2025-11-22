//go:build !windows

package executables

func fileName(name string) string {
	return name
}
