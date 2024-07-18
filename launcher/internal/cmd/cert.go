package cmd

import (
	"common"
	"fmt"
	commonExecutor "launcher-common/executor"
	"launcher/internal"
	"launcher/internal/executor"
	"launcher/internal/server"
)

func (c *Config) AddCert(canAdd string) (errorCode int) {
	if !server.CheckConnectionFromServer(common.Domain, false) {
		if canAdd != "false" {
			certMsg := fmt.Sprintf("Adding server certificate to %s store", canAdd)
			if canAdd == "user" {
				certMsg += ", accept the dialog."
			} else {
				if commonExecutor.IsAdmin() || c.StartedAgent() {
					certMsg += "."
				} else {
					certMsg += `, accept any dialog from "launcher-config-admin" if it appears.`
				}
			}
			fmt.Println(certMsg)
			var addUserCertData []byte
			var addLocalCertData []byte
			cert := server.ReadCertificateFromServer(common.Domain)
			if cert == nil {
				fmt.Println("Failed to read certificate from " + common.Domain + ". Try to access it with your browser and checking the certificate, this host must be reachable via TCP port 443 (HTTPS)")
				errorCode = internal.ErrReadCert
				return
			} else if canAdd == "local" {
				addLocalCertData = cert.Raw
			} else {
				addUserCertData = cert.Raw
			}
			if result := executor.RunSetUp(nil, addUserCertData, addLocalCertData, false, false, false); !result.Success() {
				fmt.Println("Failed to trust certificate from " + common.Domain + ".")
				errorCode = internal.ErrConfigCertAdd
				if result.Err != nil {
					fmt.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					fmt.Printf(`Exit code: %d. See documentation for "config" to check what it means.`+"\n", result.ExitCode)
				}
				return
			} else if canAdd == "local" {
				c.LocalCert()
			} else {
				c.UserCert()
			}
			if !server.CheckConnectionFromServer(common.Domain, false) {
				fmt.Println(common.Domain + " must have been trusted automatically at this point.")
				errorCode = internal.ErrServerConnectSecure
				return
			} else if !server.LanServer(common.Domain, false) {
				fmt.Println("Something went wrong, " + common.Domain + " either points to the original server or there is a certificate issue.")
				errorCode = internal.ErrTrustCert
				return
			}
		} else {
			fmt.Println(common.Domain + " must have been trusted manually. If you want it automatically, set config/option CanTrustCertificate to 'user' or 'local'.")
			errorCode = internal.ErrConfigCert
			return
		}
	} else if !server.LanServer(common.Domain, false) {
		fmt.Println("Something went wrong, " + common.Domain + " either points to the original server or there is a certificate issue.")
		errorCode = internal.ErrServerConnectSecure
	}
	return
}
