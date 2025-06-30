package executor

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"strings"
)

func EnvKey(exe string, key string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s_%s", common.Name, exe, key))
}
