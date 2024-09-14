package server

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

func HttpGet(url string, insecureSkipVerify bool) int {
	args := []string{"-f", "-s", "-4"}
	if insecureSkipVerify {
		args = append(args, "-k")
	}
	args = append(args, url)
	result := exec.Options{
		File:        "curl",
		Args:        args,
		SpecialFile: true,
		Wait:        true,
		ExitCode:    true,
	}.Exec()
	return result.ExitCode
}
