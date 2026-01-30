package gameLogs

import (
	"os"
	"path/filepath"
)

const elementSepDashGlob = "-"
const dateDashGlob = minElementGlob + minElementGlob + elementSepDashGlob + minElementGlob + elementSepDashGlob + minElementGlob
const timeDashGlob = minElementGlob + elementSepDashGlob + minElementGlob + elementSepDashGlob + minElementGlob
const dateTimeDashGlob = dateDashGlob + elementSepDotGlob + timeDashGlob

type GameAoE4 struct{}

func (g GameAoE4) Paths(path string) (paths []string) {
	possiblePaths := []string{
		"session_data.txt",
		"warnings.log",
	}
	for _, p := range possiblePaths {
		finalPath := filepath.Join(path, p)
		if f, err := os.Stat(finalPath); err == nil && !f.IsDir() {
			paths = append(paths, finalPath)
		}
	}
	logsPath := filepath.Join(path, "LogFiles")
	if f, err := os.Stat(logsPath); err != nil || !f.IsDir() {
		return
	}
	if matches, err := filepath.Glob(filepath.Join(logsPath, "unhandled."+dateTimeDashGlob+".txt")); err == nil {
		addNewestPath(
			logsPath,
			matches,
			func(info os.FileInfo) bool {
				return !info.IsDir()
			},
			&paths,
		)
	}
	return
}
