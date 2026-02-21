package cmdUtils

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func checkCertMatch(serverId uuid.UUID, gameId string, serverCertificate *x509.Certificate, hosts []string, rootCAs *x509.CertPool, fixable bool) (requiresFixing bool, errorCode int) {
	for _, host := range hosts {
		if !server.CheckConnectionFromServer(host, false, rootCAs) {
			if fixable {
				cert := server.ReadCACertificateFromServer(host)
				if cert == nil {
					logger.Println("Failed to read certificate from " + host + ".")
					errorCode = internal.ErrReadCert
					return
				} else if !bytes.Equal(cert.Raw, serverCertificate.Raw) {
					logger.Println("The certificate for " + host + " does not match the server certificate.")
					errorCode = internal.ErrCertMismatch
					return
				}
				requiresFixing = true
			} else {
				logger.Println(host + " must have been trusted manually.")
				errorCode = internal.ErrConfigCert
				return
			}
		} else if cert := server.ReadCACertificateFromServer(host); cert == nil || !bytes.Equal(cert.Raw, serverCertificate.Raw) {
			logger.Println("The certificate for " + host + " does not match the server certificate (or could not be read).")
			errorCode = internal.ErrCertMismatch
			return
		} else if !server.LanServerHost(serverId, gameId, host, false, rootCAs) {
			logger.Println("Something went wrong, " + host + " does not point to a lan server.")
			errorCode = internal.ErrServerConnectSecure
			return
		}
	}
	return
}

func (c *Config) AddCert(gameId string, serverId uuid.UUID, serverCertificate *x509.Certificate, canAdd string, customCertFile bool) (errorCode int) {
	hosts := common.AllHosts(gameId)
	var addCert bool
	if customCertFile {
		addCert = true
	} else {
		addCert, errorCode = checkCertMatch(serverId, gameId, serverCertificate, hosts, nil, canAdd != "false")
		if errorCode != common.ErrSuccess {
			return
		}
	}
	if !addCert {
		return
	}
	var certMsg string
	var addUserCertData []byte
	var addLocalCertData []byte
	if customCertFile {
		certFile, err := os.CreateTemp("", common.Name+"_cert_*.pem")
		if err != nil {
			return internal.ErrConfigCertAdd
		}
		if err = certFile.Close(); err != nil {
			return internal.ErrConfigCertAdd
		}
		c.certFilePath, _ = filepath.Abs(certFile.Name())
		addLocalCertData = serverCertificate.Raw
		certMsg = fmt.Sprintf("Saving 'server' certificate to '%s' file", certFile.Name())
	} else {
		certMsg = fmt.Sprintf("Adding 'server' certificate to %s store", canAdd)
		if canAdd == "user" {
			certMsg += ", accept the dialog"
		} else {
			if !launcherCommon.ConfigAdminAgentRunning(false) {
				certMsg += `, authorize 'config-admin-agent' if needed`
			}
		}
		if canAdd == "local" {
			addLocalCertData = serverCertificate.Raw
		} else {
			addUserCertData = serverCertificate.Raw
		}
	}
	certMsg += "..."
	logger.Println(certMsg)
	var err error
	if err = commonLogger.FileLogger.Buffer("config_setup_CA_store", func(writer io.Writer) {
		if result := executor.RunSetUp(gameId, nil, addUserCertData, addLocalCertData, nil, false, false, false, "", c.certFilePath, "", writer, func(options exec.Options) {
			commonLogger.Println("run config setup for CA store cert", options.String())
		}); !result.Success() {
			if customCertFile {
				logger.Println("Failed to save certificate to file")
			} else {
				logger.Println("Failed to trust certificate")
			}
			errorCode = internal.ErrConfigCertAdd
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			return
		}
	}); err != nil {
		return common.ErrFileLog
	}
	if !customCertFile {
		for _, host := range hosts {
			if !server.CheckConnectionFromServer(host, false, nil) {
				logger.Println(host + " must have been trusted automatically at this point.")
				errorCode = internal.ErrServerConnectSecure
				return
			} else if !server.LanServerHost(serverId, gameId, host, false, nil) {
				logger.Println("Something went wrong, " + host + " either points to the original 'server' or there is a certificate issue.")
				errorCode = internal.ErrTrustCert
				return
			}
		}
	}
	return
}
