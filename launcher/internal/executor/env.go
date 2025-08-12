package executor

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"strings"
)

func EnvMap(exe string, values map[string]string) map[string]string {
	envMap := make(map[string]string, len(values))
	for k, v := range values {
		envMap[envKey(exe, k)] = v
	}
	return envMap
}

func envKey(exe string, key string) string {
	return strings.ToUpper(
		fmt.Sprintf("%s_%s_%s", common.Name, exe, strings.ReplaceAll(key, ".", "_")),
	)
}
