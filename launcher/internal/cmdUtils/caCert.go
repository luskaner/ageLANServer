package cmdUtils

import (
	"crypto/x509"
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) AddCACertToGame(gameId string, cert *x509.Certificate, gamePath string) (errorCode int) {
	logger.Println("Adding CA certificate to game...")
	var err error
	if err = commonLogger.FileLogger.Buffer("config_setup_CA_game", func(writer io.Writer) {
		if result := executor.RunSetUp(gameId, nil, nil, nil, cert.Raw, false, false, false, false, "", "", gamePath, writer, func(options exec.Options) {
			commonLogger.Println("run config setup for CA game cert", options.String())
		}); !result.Success() {
			logger.Println("Failed to save CA certificate to game")
			errorCode = internal.ErrConfigCACertAdd
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		}
	}); err != nil {
		logger.Println("Error message: " + err.Error())
		return common.ErrFileLog
	}
	return
}
