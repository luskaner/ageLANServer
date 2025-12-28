package cmdUtils

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

var reWinToLinVar = regexp.MustCompile(`%(\w+)%`)

func ParseCommandArgs(cmdSlice []string, values map[string]string) (args []string, err error) {
	cmdArgs := strings.Join(cmdSlice, " ")
	for key, value := range values {
		cmdArgs = strings.ReplaceAll(cmdArgs, fmt.Sprintf(`{%s}`, key), value)
	}
	if runtime.GOOS == "windows" {
		cmdArgs = reWinToLinVar.ReplaceAllString(cmdArgs, `$$$1`)
	}
	args, err = shell.Fields(cmdArgs, nil)
	return
}
