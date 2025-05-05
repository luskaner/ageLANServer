package cmdUtils

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func (c *Config) AddCert(serverCertificate *x509.Certificate, canAdd string) (errorCode int) {
	hosts := common.AllHosts()
	var addCert bool
	for _, host := range hosts {
		if !server.CheckConnectionFromServer(host, false) {
			if canAdd != "false" {
				cert := server.ReadCertificateFromServer(host)
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
		} else if cert := server.ReadCertificateFromServer(host); cert == nil || !bytes.Equal(cert.Raw, serverCertificate.Raw) {
			fmt.Println("The certificate for " + host + " does not match the server certificate (or could not be read).")
			errorCode = internal.ErrCertMismatch
			return
		} else if !server.LanServer(host, false) {
			fmt.Println("Something went wrong, " + host + " does not point to a lan server.")
			errorCode = internal.ErrServerConnectSecure
			return
		}
	}
	if !addCert {
		return
	}
	certMsg := fmt.Sprintf("Adding 'server' certificate to %s store", canAdd)
	if canAdd == "user" {
		certMsg += ", accept the dialog."
	} else {
		if !c.CfgAgentStarted() {
			certMsg += `, authorize 'config-admin-agent' if needed.`
		}
	}
	fmt.Println(certMsg)
	var addUserCertData []byte
	var addLocalCertData []byte
	if canAdd == "local" {
		addLocalCertData = serverCertificate.Raw
	} else {
		addUserCertData = serverCertificate.Raw
	}
	if result := executor.RunSetUp("", nil, addUserCertData, addLocalCertData, false, false, false, false, ""); !result.Success() {
		fmt.Println("Failed to trust certificate")
		errorCode = internal.ErrConfigCertAdd
		if result.Err != nil {
			fmt.Println("Error message: " + result.Err.Error())
		}
		if result.ExitCode != common.ErrSuccess {
			fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
		}
		return
	} else if canAdd == "local" {
		c.LocalCert()
	} else {
		c.UserCert()
	}
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
	return
}
