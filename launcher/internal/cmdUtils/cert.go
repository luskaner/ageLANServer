package cmdUtils

import (
	"bytes"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
)

func (c *Config) AddCert(canAdd string) (errorCode int) {
	var previousCert []byte
	var previousHost string
	hosts := common.AllHosts()
	for _, host := range hosts {
		if !server.CheckConnectionFromServer(host, false) {
			if canAdd != "false" {
				cert := server.ReadCertificateFromServer(host)
				if cert == nil {
					fmt.Println("Failed to read certificate from " + host + ". Try to access it with your browser and checking the certificate, this host must be reachable via TCP port 443 (HTTPS)")
					errorCode = internal.ErrReadCert
					return
				} else if len(previousCert) > 0 && !bytes.Equal(previousCert, cert.Raw) {
					fmt.Println("The certificate for " + host + " does not match the previous for " + previousHost)
					errorCode = internal.ErrCertMismatch
					return
				}
				previousCert = cert.Raw
				previousHost = host
			} else {
				fmt.Println(host + " must have been trusted manually. If you want it automatically, set config/option CanTrustCertificate to 'user' or 'local'.")
				errorCode = internal.ErrConfigCert
				return
			}
		} else if !server.LanServer(host, false) {
			fmt.Println("Something went wrong, " + host + " points to the original 'server'.")
			errorCode = internal.ErrServerConnectSecure
			return
		}
	}
	if len(previousCert) == 0 {
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
		addLocalCertData = previousCert
	} else {
		addUserCertData = previousCert
	}
	if result := executor.RunSetUp("", nil, addUserCertData, addLocalCertData, false, false, false, false); !result.Success() {
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
