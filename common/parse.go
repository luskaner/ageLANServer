package common

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/viper"
	"mvdan.cc/sh/v3/shell"
)

var reWinToLinVar *regexp.Regexp

func ParseCommandArgs(name string, values map[string]string, separateFields bool) (args []string, err error) {
	cmdArgs := strings.Join(viper.GetStringSlice(name), " ")
	for key, value := range values {
		cmdArgs = strings.ReplaceAll(cmdArgs, fmt.Sprintf(`{%s}`, key), value)
	}
	if runtime.GOOS == "windows" {
		if reWinToLinVar == nil {
			reWinToLinVar = regexp.MustCompile(`%(\w+)%`)
		}
		cmdArgs = reWinToLinVar.ReplaceAllString(cmdArgs, `$$$1`)
	}
	if separateFields {
		args, err = shell.Fields(cmdArgs, nil)
	} else {
		args = []string{cmdArgs}
	}
	return
}

func ParsePath(name string, values map[string]string) (file os.FileInfo, path string, err error) {
	args, err := ParseCommandArgs(name, values, false)
	if err != nil {
		return
	}
	if len(args) != 1 {
		err = fmt.Errorf("invalid path")
		return
	}
	path = args[0]
	file, err = os.Stat(args[0])
	return
}
