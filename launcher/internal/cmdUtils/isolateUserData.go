package cmdUtils

import (
	"io"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func ResolveIsolateValue(value string, officialLauncher bool) bool {
	switch value {
	case "true":
		return true
	case "false":
		return false
	case "required":
		return officialLauncher
	default:
		return false
	}
}

func (c *Config) IsolateUserData(metadata bool, profiles bool, path string) (errorCode int) {
	if metadata || profiles {
		var isolateItems []string
		if metadata {
			isolateItems = append(isolateItems, "metadata")
		}
		if profiles {
			isolateItems = append(isolateItems, "profiles")
		}
		logger.Println("Backing up " + strings.Join(isolateItems, " and ") + ".")
		var err error
		if err = commonLogger.FileLogger.Buffer("config_setup_isolate", func(writer io.Writer) {
			cfgSetupOpts := &executor.ConfigSetupOptions{
				GameId:         c.gameId,
				BackupMetadata: metadata,
				BackupProfiles: profiles,
				GameDataPath:   path,
				Out:            writer,
				OptionsFn: func(options exec.Options) {
					commonLogger.Println("run config setup for data isolation", options.String())
				},
			}
			if result := cfgSetupOpts.RunSetUp(); !result.Success() {
				isolateMsg := "Failed to backup "
				logger.Println(isolateMsg + strings.Join(isolateItems, " or ") + ".")
				errorCode = internal.ErrMetadataProfilesSetup
				if result.Err != nil {
					logger.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
				}
			}
		}); err != nil {
			return common.ErrFileLog
		}
	}
	return
}
