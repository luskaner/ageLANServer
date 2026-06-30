package executor

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
)

func ExecuteBattleServer(gameId string, path string, region string, name string, ports []int, certFile string,
	keyFile string, extraArgs []string, hideWindow bool, logRoot string) (pid uint32, err error) {
	var simulationPeriod int
	switch gameId {
	case game.AoE1:
		simulationPeriod = 25
	case game.AoE3, game.AoM:
		simulationPeriod = 50
	case game.AoE2, game.AoE4:
		simulationPeriod = 125
	}
	bsPort := fmt.Sprintf("%d", ports[0])
	args := []string{
		"-region", region,
		"-name", name,
		"-publicPort", bsPort,
		"-relaybroadcastPort", "0",
		"-simulationPeriod", fmt.Sprintf("%d", simulationPeriod),
		"-bsPort", bsPort,
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
