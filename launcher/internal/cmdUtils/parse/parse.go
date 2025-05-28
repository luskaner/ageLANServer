package parse

import (
	"fmt"
	"mvdan.cc/sh/v3/shell"
	"strings"
)

func insertValues(cmdArgs string, values map[string]string) string {
	for key, value := range values {
		cmdArgs = strings.ReplaceAll(cmdArgs, fmt.Sprintf(`{%s}`, key), value)
	}
	return cmdArgs
}

func CommandArgs(value []string, values map[string]string) (args []string, err error) {
	cmdArgs := strings.Join(value, " ")
	cmdArgs = insertValues(cmdArgs, values)
	cmdArgs = postprocess(cmdArgs)
	args, err = shell.Fields(cmdArgs, nil)
	return
}

func Executable(value string, values map[string]string) (exe string, err error) {
	value = insertValues(value, values)
	value = postprocess(value)
	exe, err = shell.Expand(value, nil)
	return
}
