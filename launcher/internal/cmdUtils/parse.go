package cmdUtils

import (
	"fmt"
	"os"
	"strings"
)

func insertValues(cmdArgs string, values map[string]string) string {
	for key, value := range values {
		cmdArgs = strings.ReplaceAll(cmdArgs, fmt.Sprintf(`{%s}`, key), value)
	}
	return cmdArgs
}

func CommandArgs(args []string, values map[string]string) (argsTransformed []string) {
	argsTransformed = make([]string, len(args))
	for i := range args {
		argsTransformed[i] = os.ExpandEnv(insertValues(args[i], values))
	}
	return
}
