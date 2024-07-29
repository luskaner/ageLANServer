package game

import (
	commonProcess "common/process"
	mapset "github.com/deckarep/golang-set/v2"
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
	return commonProcess.AnyProcessExists(processes.ToSlice())
}
