//go:build !windows

package serverKill

import commonProcess "github.com/luskaner/ageLANServer/common/process"

func Do(path string) error {
	_, err := commonProcess.Kill(path)
	return err
}
