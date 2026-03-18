//go:build !windows

package hosts

import "math"

const maxHostsPerLine = math.MaxInt // Not an actual limit
const maxCharsPerLine = math.MaxUint8 + 1
