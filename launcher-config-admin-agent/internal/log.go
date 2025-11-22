package internal

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
)

func InitializeOrExit(logRoot string) {
	if err := commonLogger.NewOwnFileLogger("config-admin-agent", logRoot, "", true); err != nil {
		os.Exit(common.ErrFileLog)
	}
}
