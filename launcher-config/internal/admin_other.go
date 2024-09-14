//go:build !windows

package internal

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common/process"
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"time"
)

func preAgentStart() {
	if !exec.IsAdmin() {
		fmt.Println("Waiting up to 30s for agent to start...")
	}
}

func postAgentStart(file string) {
	if !exec.IsAdmin() {
		for i := 0; i < 30; i++ {
			_, proc, _ := process.Process(file)
			if proc != nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
}
