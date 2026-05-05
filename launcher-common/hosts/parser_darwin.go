package hosts

import (
	"unicode"
)

func splitLine(line string) []string {
	var prevWasSpace bool
	for i, r := range line {
		if r == commentMarker && prevWasSpace {
			return []string{line[:i], line[i+1:]}
		}
		prevWasSpace = unicode.IsSpace(r)
	}
	return []string{line}
}
