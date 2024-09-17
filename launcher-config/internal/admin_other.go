//go:build !windows

package internal

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common/executor"
	"github.com/luskaner/aoe2DELanServer/common/process"
	"time"
)

func preAgentStart() {
	if !executor.IsAdmin() {
		fmt.Println("Waiting up to 30s for agent to start...")
	}
}

func postAgentStart(file string) {
	if !executor.IsAdmin() {
		for i := 0; i < 30; i++ {
			_, proc, _ := process.Process(file)
			if proc != nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
}
