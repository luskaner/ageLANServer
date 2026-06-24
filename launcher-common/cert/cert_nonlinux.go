//go:build !linux

package cert

import "github.com/luskaner/ageLANServer/common/executor/exec"

func FlushCerts() (result *exec.Result) {
	return exec.ResultSuccess
}
