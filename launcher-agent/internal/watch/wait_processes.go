package watch

import (
	"os"
	"sort"

	commonProcess "github.com/luskaner/ageLANServer/common/process"
)

// sortedProcessNames returns the map keys in a deterministic order so log
// output and waiting order do not depend on Go's randomized map iteration.
func sortedProcessNames(processes map[string]*os.Process) []string {
	names := make([]string, 0, len(processes))
	for name := range processes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// waitForProcessesToExit waits for every process to exit. It used to wait for
// a single randomly-chosen process: if that one exited first while the game
// was still running, the agent reverted the configuration too early.
func waitForProcessesToExit(processes []*os.Process) bool {
	allExited := true
	for _, p := range processes {
		if p == nil {
			continue
		}
		if !commonProcess.WaitForProcess(p, nil) {
			allExited = false
		}
	}
	return allExited
}
