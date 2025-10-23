package cmdUtils

import (
	"fmt"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) IsolateUserData(metadata bool, profiles bool) (errorCode int) {
	if metadata || profiles {
		var isolateItems []string
		if metadata {
			isolateItems = append(isolateItems, "metadata")
		}
		if profiles {
			isolateItems = append(isolateItems, "profiles")
		}
		fmt.Println("Backing up " + strings.Join(isolateItems, " and ") + ".")
		if result := executor.RunSetUp(c.gameId, nil, nil, nil, nil, metadata, profiles, false, false, "", "", "", func(options exec.Options) {
			LogPrintln("run config setup for data isolation", options.String())
		}); !result.Success() {
			isolateMsg := "Failed to backup "
			fmt.Println(isolateMsg + strings.Join(isolateItems, " or ") + ".")
			errorCode = internal.ErrMetadataProfilesSetup
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		}
	}
	return
}
