//go:build !darwin

package hosts

import "strings"

func splitLine(line string) (splitted []string) {
	return strings.SplitN(line, string(commentMarker), 2)
}
