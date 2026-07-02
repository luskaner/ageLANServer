//go:build !darwin

package steam

func NewExec(gameId string) (exec *Exec, ok bool) {
	return newExec(gameId)
}

func (exec Exec) GameProcesses() (steamProcess bool, steamMacOsNative bool, xboxProcess bool) {
	steamProcess = true
	return
}
