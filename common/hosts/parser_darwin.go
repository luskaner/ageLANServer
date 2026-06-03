package hosts

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const markerStr = string(commentMarker)

func splitLine(line string) []string {
	if strings.HasPrefix(line, markerStr) {
		return []string{"", line[len(markerStr):]}
	}
	var prevWasSpace bool
	for i, r := range line {
		if r == commentMarker && prevWasSpace {
			cut := i + utf8.RuneLen(r)
			return []string{line[:i], line[cut:]}
		}
		prevWasSpace = unicode.IsSpace(r)
	}
	return []string{line}
}
