package cmdUtils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/appx"
	"github.com/luskaner/ageLANServer/common/game/steam"
	"github.com/luskaner/ageLANServer/common/logger"
)

func ResolvePath(gameId string, executablePath string) (resolvedPath string, err error) {
	validPath := func(path string) bool {
		if f, localErr := os.Stat(path); localErr == nil && !f.IsDir() {
			return true
		}
		return false
	}
	var path string
	if executablePath == "auto" {
		commonLogger.Println("Auto resolving executable path...")
		// TODO: Review if AoE: DE and AoE III: DE can also use AoE II: DE
		// The Battle Server for AoM is buggy and the only one working is the AoE II one
		if gameId == common.GameAoM {
			gameId = common.GameAoE2
		}
		battleServerPath := "BattleServer.exe"
		if gameId == common.GameAoE2 {
			battleServerPath = filepath.Join("BattleServer", battleServerPath)
		}
		game := steam.NewGame(gameId)
		libraryFolder := game.LibraryFolder()
		if libraryFolder != "" {
			folder := game.Path(libraryFolder)
			if folder != "" {
				path = filepath.Join(folder, battleServerPath)
				if validPath(path) {
					commonLogger.Println("\tFound in Steam")
					return path, nil
				}
			}
		}
		if ok, folder := appx.GameInstallLocation(gameId); ok {
			path = filepath.Join(folder, battleServerPath)
			if validPath(path) {
				commonLogger.Println("\tFound on Xbox")
				return path, nil
			}
		}
		err = fmt.Errorf("could not find battle server executable")
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
	case common.GameAoE2:
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
	if result := options.Exec(); result.Success() {
		pid = result.Pid
	} else {
		err = result.Err
	}
	return
}
