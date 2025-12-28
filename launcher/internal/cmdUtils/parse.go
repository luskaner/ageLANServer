package cmdUtils

import (
	"github.com/luskaner/ageLANServer/common"
)

func ParseCommandArgs(cmdSlice []string, values map[string]string) (args []string, err error) {
	return common.ParseCommandArgsFromSlice(cmdSlice, values, true)
}
