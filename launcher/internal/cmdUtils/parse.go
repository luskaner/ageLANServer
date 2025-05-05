package cmdUtils

import (
	"fmt"
	"github.com/spf13/viper"
	"mvdan.cc/sh/v3/shell"
	"regexp"
	"runtime"
	"strings"
)

var reWinToLinVar = regexp.MustCompile(`%(\w+)%`)

func ParseCommandArgs(name string, values map[string]string) (args []string, err error) {
	cmdArgs := strings.Join(viper.GetStringSlice(name), " ")
	for key, value := range values {
		cmdArgs = strings.ReplaceAll(cmdArgs, fmt.Sprintf(`{%s}`, key), value)
	}
	if runtime.GOOS == "windows" {
		cmdArgs = reWinToLinVar.ReplaceAllString(cmdArgs, `$$$1`)
	}
	args, err = shell.Fields(cmdArgs, nil)
	return
}
