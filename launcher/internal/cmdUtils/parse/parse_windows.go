package parse

import (
	"regexp"
)

var reWinToLinVar *regexp.Regexp

func postprocess(value string) string {
	if reWinToLinVar == nil {
		reWinToLinVar = regexp.MustCompile(`%(\w+)%`)
	}
	return reWinToLinVar.ReplaceAllString(value, `$$$1`)
}
