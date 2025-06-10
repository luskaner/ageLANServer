package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) IsolateUserData(windowsUserProfilePath string, metadata bool) (errorCode int) {
	fmt.Print(printer.Gen(
		printer.Configuration,
		"",
		"Backing up metadata... ",
	))
	if result := executor.RunSetUp(&executor.RunSetUpOptions{Game: c.gameId, WindowsUserProfilePath: windowsUserProfilePath, BackupMetadata: metadata}); !result.Success() {
		printer.PrintFailedResultError(result)
		errorCode = internal.ErrMetadataSetup
	} else {
		printer.PrintSucceeded()
	}
	return
}
