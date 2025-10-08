package cmdUtils

import (
	"time"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
)

func WaitForBattleServerInit(config battleServerConfig.Config) (ok bool) {
	// Wait up to 10s to initialize
	timeout := time.After(10 * time.Second)
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
