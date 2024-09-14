package hosts

import "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"

func flushDns() (result *exec.Result) {
	result = exec.Options{File: "dscacheutil", ExitCode: true, Wait: true, Args: []string{"-flushcache", "&&", "killall", "-HUP", "mDNSResponder"}}.Exec()
	return
}
