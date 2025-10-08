package cmdUtils

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func (c *Config) AddCert(gameId string, serverCertificate *x509.Certificate, canAdd string, customCertFile bool) (errorCode int) {
	hosts := common.AllHosts(gameId)
	var addCert bool
	if customCertFile {
		addCert = true
	} else {
		for _, host := range hosts {
			if !server.CheckConnectionFromServer(host, false) {
				if canAdd != "false" {
					cert := server.ReadCACertificateFromServer(host)
					if cert == nil {
						fmt.Println("Failed to read certificate from " + host + ".")
						errorCode = internal.ErrReadCert
						return
					} else if !bytes.Equal(cert.Raw, serverCertificate.Raw) {
						fmt.Println("The certificate for " + host + " does not match the server certificate.")
						errorCode = internal.ErrCertMismatch
						return
					}
					addCert = true
				} else {
					fmt.Println(host + " must have been trusted manually. If you want it automatically, set config/option CanTrustCertificate to 'user' or 'local'.")
					errorCode = internal.ErrConfigCert
					return
				}
			} else if cert := server.ReadCACertificateFromServer(host); cert == nil || !bytes.Equal(cert.Raw, serverCertificate.Raw) {
				fmt.Println("The certificate for " + host + " does not match the server certificate (or could not be read).")
				errorCode = internal.ErrCertMismatch
				return
			} else if !server.LanServer(host, false) {
				fmt.Println("Something went wrong, " + host + " does not point to a lan server.")
				errorCode = internal.ErrServerConnectSecure
				return
			}
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
		c.certFilePath = certFile.Name()
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
	fmt.Println(certMsg)
	if result := executor.RunSetUp(gameId, nil, addUserCertData, addLocalCertData, nil, false, false, false, false, "", c.certFilePath, ""); !result.Success() {
		if customCertFile {
			fmt.Println("Failed to save certificate to file")
		} else {
			fmt.Println("Failed to trust certificate")
		}
		errorCode = internal.ErrConfigCertAdd
		if result.Err != nil {
			fmt.Println("Error message: " + result.Err.Error())
		}
		if result.ExitCode != common.ErrSuccess {
			fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
		}
		return
	}
	if !customCertFile {
		for _, host := range hosts {
			if !server.CheckConnectionFromServer(host, false) {
				fmt.Println(host + " must have been trusted automatically at this point.")
				errorCode = internal.ErrServerConnectSecure
				return
			} else if !server.LanServer(host, false) {
				fmt.Println("Something went wrong, " + host + " either points to the original 'server' or there is a certificate issue.")
				errorCode = internal.ErrTrustCert
				return
			}
		}
	}
	return
}
