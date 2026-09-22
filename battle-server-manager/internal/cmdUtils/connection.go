package cmdUtils

import (
	"runtime"
	"time"

	"github.com/luskaner/ageLANServer/common/battleServer"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
)

var waitInitTimeout = 10 * time.Second

func WaitForBattleServerInit(config battleServer.Config) (ok bool) {
	// Wait for initialization
	t := waitInitTimeout
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
			if ok = config.Validate(false); ok {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return
}
