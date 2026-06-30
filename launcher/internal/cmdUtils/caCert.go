package cmdUtils

import (
	"crypto/x509"
	"io"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) AddCACertToGame(gameId string, serverId uuid.UUID, serverCertificate *x509.Certificate, gamePath string, caCertPath string, canAddCert bool) (exitCode int) {
	logger.Println("Adding CA certificate to game if needed...")
	caPool, err := common.ReadCertsPool(caCertPath)
	if err != nil {
		logger.Println("Could not read game CA certificates:", err)
		return internal.ErrConfigCACertAdd
	}
	var addCert bool
	addCert, exitCode = checkCertMatch(serverId, gameId, serverCertificate, common.AllHosts(gameId), caPool, canAddCert)
	if !addCert || exitCode != common.ErrSuccess {
		return
	}
	if err = commonLogger.FileLogger.Buffer("config_setup_CA_game", func(writer io.Writer) {
		cfgSetupOpts := executor.NewConfigSetupOptions()
		cfgSetupOpts.Out = writer
		cfgSetupOpts.OptionsFn = func(options exec.Options) {
			commonLogger.Println("run config setup for CA game cert", options.String())
		}
		cfgSetupOpts.GameId = gameId
		cfgSetupOpts.GamePath = gamePath
		cfgSetupOpts.AddCACertData = serverCertificate.Raw
		if result := cfgSetupOpts.RunSetUp(); !result.Success() {
			logger.Println("Failed to save CA certificate to game")
			exitCode = internal.ErrConfigCACertAdd
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
