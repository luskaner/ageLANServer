package cmdUtils

import (
	"crypto/x509"
	"io"

	"github.com/luskaner/ageLANServer/common/uuid"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) AddCACertToGame(gameId string, serverId uuid.UUID, serverCertificate *x509.Certificate, gamePath string, caCertPath string, canAddCert bool, macOsExclusiveMappings bool) (exitCode int) {
	logger.Step("Adding CA certificate to game if needed...")
	caPool, err := common.ReadCertsPool(caCertPath)
	if err != nil {
		logger.Fail("Could not read game CA certificates: %s", err)
		return internal.ErrConfigCACertAdd
	}
	var addCert bool
	addCert, exitCode = checkCertMatch(serverId, gameId, serverCertificate, common.AllHosts(gameId, macOsExclusiveMappings), caPool, canAddCert)
	if !addCert || exitCode != common.ErrSuccess {
		return
	}
	if err = commonLogger.FileLogger.Buffer("config_setup_CA_game", func(writer io.Writer) {
		cfgSetupOpts := executor.NewConfigSetupOptions()
		cfgSetupOpts.Out = writer
		cfgSetupOpts.OptionsFn = func(options *exec.Options) {
			commonLogger.Println("run config setup for CA game cert", options.String())
		}
		cfgSetupOpts.GameId = gameId
		cfgSetupOpts.GamePath = gamePath
		cfgSetupOpts.AddCACertData = serverCertificate.Raw
		cfgSetupOpts.AgentEndOnError = !c.RequiresConfigRevert()
		if result := cfgSetupOpts.RunSetUp(); !result.Success() {
			logger.Fail("Failed to save CA certificate to game")
			exitCode = internal.ErrConfigCACertAdd
			if result.Err != nil {
				logger.Fault("Error message: %s", result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Fault("Exit code: %d.", result.ExitCode)
			}
		}
	}); err != nil {
		logger.Fail("Error message: %s", err.Error())
		return common.ErrFileLog
	}
	return
}
