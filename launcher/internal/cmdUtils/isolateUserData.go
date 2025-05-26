package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) IsolateUserData(windowsUserProfilePath string, metadata bool) (errorCode int) {
	fmt.Println("Backing up metadata.")
	if result := executor.RunSetUp(&executor.RunSetUpOptions{Game: c.gameId, WindowsUserProfilePath: windowsUserProfilePath, BackupMetadata: metadata}); !result.Success() {
		fmt.Println("Failed to backup metadata.")
		errorCode = internal.ErrMetadataSetup
		if result.Err != nil {
			fmt.Println("Error message: " + result.Err.Error())
		}
		if result.ExitCode != common.ErrSuccess {
			fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
		}
	}
	return
}
