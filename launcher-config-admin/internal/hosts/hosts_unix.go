//go:build !windows

package hosts

const LineEnding = "\n"

func Path() string {
	return "/etc/hosts"
}
