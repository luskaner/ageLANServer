package hosts

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`\s#$`)

func splitLine(line string) (splitted []string) {
	presplit := strings.SplitAfter(line, string(commentMarker))
	for i := 0; i < len(presplit)-1; i++ {
		if re.MatchString(presplit[i]) {
			presplit[i] = presplit[i][:len(presplit[i])-2]
		} else {
			presplit[i+1] = presplit[i] + presplit[i+1]
			presplit[i] = ""
		}
	}
	for _, s := range presplit {
		if s != "" {
			splitted = append(splitted, s)
		}
	}
	if len(splitted) == 0 {
		splitted = append(splitted, "")
	}
	return
}
