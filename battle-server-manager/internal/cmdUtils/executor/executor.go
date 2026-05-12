package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
)

func validPath(path string) bool {
	if f, localErr := os.Stat(path); localErr == nil && !f.IsDir() {
		return true
	}
	return false
}

func locatablePath(locFn func(gameId string) (game game.Locatable, ok bool), gameId string, battleServerPath string, name string) (path string) {
	if locatable, ok := locFn(gameId); ok {
		if folder := locatable.Path(); folder != "" {
			tmpPath := filepath.Join(folder, battleServerPath)
			if validPath(tmpPath) {
				path = tmpPath
				commonLogger.Printf("\tFound in %s\n", name)
			}
		}
	}
	return
}

func ResolvePath(gameId string, executablePath string) (resolvedPath string, err error) {
	var path string
	if executablePath == "auto" {
		commonLogger.Println("Auto resolving executable path...")
		// TODO: Review if AoE: DE and AoE III: DE can also use AoE II: DE
		// The Battle Server for AoM is buggy and the only one working is the AoE II one.
		// AoE IV does not have one.
		if gameId == common.GameAoE4 || gameId == common.GameAoM {
			gameId = common.GameAoE2
		}
		battleServerPath := "BattleServer.exe"
		if gameId == common.GameAoE2 {
			battleServerPath = filepath.Join("BattleServer", battleServerPath)
		}
		if path = resolveAutoPath(gameId, battleServerPath); path == "" {
			err = fmt.Errorf("could not find battle server executable")
		} else if validPath(path) {
			resolvedPath = path
		}
		return
	}
	var pathErr error
	_, path, pathErr = common.ParsePath(common.EnhancedViperStringToStringSlice(executablePath), nil)
	if pathErr != nil {
		err = fmt.Errorf("invalid battle server executable path")
		return
	}
	if validPath(path) {
		resolvedPath = path
	} else {
		err = fmt.Errorf("invalid battle server executable path")
	}
	return
}

func ExecuteBattleServer(gameId string, path string, region string, name string, ports []int, certFile string,
	keyFile string, extraArgs []string, hideWindow bool, logRoot string) (pid uint32, err error) {
	var simulationPeriod int
	switch gameId {
	case common.GameAoE1:
		simulationPeriod = 25
	case common.GameAoE3, common.GameAoM:
		simulationPeriod = 50
	case common.GameAoE2, common.GameAoE4:
		simulationPeriod = 125
	}
	args := []string{
		"-region", region,
		"-name", name,
		"-relaybroadcastPort", "0",
		"-simulationPeriod", fmt.Sprintf("%d", simulationPeriod),
		"-bsPort", fmt.Sprintf("%d", ports[0]),
		"-webSocketPort", fmt.Sprintf("%d", ports[1]),
		"-sslCert", certFile,
		"-sslKey", keyFile,
	}
	if ports[2] != -1 {
		args = append(args, "-outOfBandPort", fmt.Sprintf("%d", ports[2]))
	}
	args = append(args, extraArgs...)
	options := exec.Options{
		File:           path,
		UseWorkingPath: true,
		ShowWindow:     !hideWindow,
		Args:           args,
		Pid:            true,
	}
	if hideWindow && logRoot != "" {
		var f *os.File
		if _, f, err = commonLogger.NewFileLogger("battle-server", logRoot, "", true); err != nil {
			return
		} else if f != nil {
			options.Stdout = f
			options.Stderr = f
		}
	}
	commonLogger.Println("Executing:", options)
	if result := execWithOptions(gameId, &options); result.Success() {
		pid = result.Pid
	} else {
		err = result.Err
	}
	return
}
