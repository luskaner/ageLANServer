//go:build !windows

package hosts

import (
	"os"
)

const LineEnding = "\n"

func Path() string {
	return "/etc/hosts"
}

var hostname Host

func commentHost(host Host) bool {
	if hostname == "" {
		if name, err := os.Hostname(); err == nil {
			_, hostname = parseHost(name)
		}
	}
	return host == hostname
}
