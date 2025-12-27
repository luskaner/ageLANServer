package cmdUtils

import (
	"crypto/x509"
	"io"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommonCert "github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func readCertsPool(path string) (pool *x509.CertPool, err error) {
	var caCerts []*x509.Certificate
	_, _, caCerts, err = launcherCommonCert.ReadFromFile(path)
	if err != nil {
		return
	}
	pool = x509.NewCertPool()
	for _, caCert := range caCerts {
		pool.AddCert(caCert)
	}
	return
}

func (c *Config) AddCACertToGame(gameId string, serverId uuid.UUID, serverCertificate *x509.Certificate, gamePath string, caCertPath string) (errorCode int) {
	logger.Println("Adding CA certificate to game if needed...")
	caPool, err := readCertsPool(caCertPath)
	if err != nil {
		logger.Println("Could not read game CA certificates:", err)
	}
	var addCert bool
	addCert, errorCode = checkCertMatch(serverId, gameId, serverCertificate, common.AllHosts(gameId), caPool, true)
	if !addCert || errorCode != common.ErrSuccess {
		return
	}
	if err = commonLogger.FileLogger.Buffer("config_setup_CA_game", func(writer io.Writer) {
		if result := executor.RunSetUp(gameId, nil, nil, nil, serverCertificate.Raw, false, false, false, "", "", gamePath, writer, func(options exec.Options) {
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
