package cmdUtils

import (
	"runtime"
	"time"

	"github.com/luskaner/ageLANServer/common/battleServer"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
)

func WaitForBattleServerInit(config battleServer.Config) (ok bool) {
	// Wait for initialization
	t := 10 * time.Second
	if runtime.GOOS != "windows" {
		t *= 3
	}
	timeout := time.After(t)
	commonLogger.Printf("Waiting up to %s for the initialization to complete...", t)
loop:
	for {
		select {
		case <-timeout:
			break loop
		default:
			if ok = config.Validate(); ok {
				return
			}
		}
	}
	return
}
