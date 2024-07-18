package game

import (
	mapset "github.com/deckarep/golang-set/v2"
	"launcher-common/executor"
)

const steamProcess = "AoE2DE_s.exe"
const microsoftStoreProcess = "AoE2DE.exe"

func AnyProcessExists(steam bool, microsoftStore bool) bool {
	processes := mapset.NewSet[string]()
	if steam {
		processes.Add(steamProcess)
	}
	if microsoftStore {
		processes.Add(microsoftStoreProcess)
	}
	return executor.AnyProcessExists(processes.ToSlice())
}
