//go:build !windows

package hosts

import (
	"math"
)

const LineEnding = "\n"
const maxHostsPerLine = math.MaxInt32 - 1

func Path() string {
	return "/etc/hosts"
}
