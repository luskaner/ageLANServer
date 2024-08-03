package cmdUtils

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/executor"
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
		if result := executor.RunSetUp(nil, nil, nil, metadata, profiles, false); !result.Success() {
			isolateMsg := "Failed to backup "
			fmt.Println(isolateMsg + strings.Join(isolateItems, " or ") + ".")
			errorCode = internal.ErrMetadataProfilesSetup
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "config" to check what it means.`+"\n", result.ExitCode)
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
