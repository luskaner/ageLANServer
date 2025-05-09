package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"strings"
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
		if result := executor.RunSetUp(c.gameId, nil, nil, nil, metadata, profiles, false, false, "", ""); !result.Success() {
			isolateMsg := "Failed to backup "
			fmt.Println(isolateMsg + strings.Join(isolateItems, " or ") + ".")
			errorCode = internal.ErrMetadataProfilesSetup
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		} else {
			if metadata {
				c.BackedUpMetadata()
			}
			if profiles {
				c.BackedUpProfiles()
			}
		}
	}
	return
}
