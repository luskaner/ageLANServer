package cmdUtils

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server"
	"os"
	"runtime"
)

func (c *Config) AddCert(serverId uuid.UUID, serverCertificate *x509.Certificate, store string) (errorCode int) {
	hosts := common.AllHosts()
	var addCert bool
	if store == "tmp" {
		addCert = true
	} else {
		for _, host := range hosts {
			if !server.CheckConnectionFromServer(host, false) {
				if store != "" {
					cert := server.ReadCertificateFromServer(host)
					if cert == nil {
						printer.Println(
							printer.Error,
							printer.T("Failed to read certificate from "),
							printer.TS(host, printer.LiteralStyle),
							printer.T("."),
						)
						errorCode = internal.ErrReadCert
						return
					} else if !bytes.Equal(cert.Raw, serverCertificate.Raw) {
						printer.Println(
							printer.Error,
							printer.T("The certificate for "),
							printer.TS(host, printer.LiteralStyle),
							printer.T(" does not match the "),
							printer.TS("server", printer.ComponentStyle),
							printer.T(" certificate."),
						)
						errorCode = internal.ErrCertMismatch
						return
					}
					addCert = true
				} else {
					styledTexts := []*printer.StyledText{
						printer.TS(host, printer.LiteralStyle),
						printer.T(" must have been trusted manually. If you want it automatically, set "),
						printer.TS("CanTrustCertificate", printer.LiteralStyle),
						printer.TS("local", printer.LiteralStyle),
					}
					if runtime.GOOS == "windows" {
						styledTexts = append(
							styledTexts,
							printer.T(" or "),
							printer.TS("user", printer.LiteralStyle),
						)
					}
					styledTexts = append(styledTexts, printer.T("."))
					printer.Println(
						printer.Error,
						styledTexts...,
					)
					errorCode = internal.ErrConfigCert
					return
				}
			} else if cert := server.ReadCertificateFromServer(host); cert == nil || !bytes.Equal(cert.Raw, serverCertificate.Raw) {
				printer.Println(
					printer.Error,
					printer.T("The certificate for "),
					printer.TS(host, printer.LiteralStyle),
					printer.T(" does not match the "),
					printer.TS("server", printer.ComponentStyle),
					printer.T(" certificate (or could not be read)."),
				)
				errorCode = internal.ErrCertMismatch
				return
			} else if !server.LanServerHost(serverId, c.gameTitle, host, false) {
				printer.Println(
					printer.Error,
					printer.T("Something went wrong, "),
					printer.TS(host, printer.LiteralStyle),
					printer.T(" does not point to the LAN "),
					printer.TS("server", printer.ComponentStyle),
					printer.T("."),
				)
				errorCode = internal.ErrServerConnectSecure
				return
			}
		}
	}
	if !addCert {
		return
	}
	var certStyledTexts []*printer.StyledText
	var addUserCertData []byte
	var addLocalCertData []byte
	if store == "tmp" {
		certFile, err := os.CreateTemp("", common.Name+"_cert_*.pem")
		if err != nil {
			return internal.ErrConfigCertAdd
		}
		if err = certFile.Close(); err != nil {
			return internal.ErrConfigCertAdd
		}
		c.SetCertFilePath(certFile.Name())
		addLocalCertData = serverCertificate.Raw
		certStyledTexts = append(
			certStyledTexts,
			printer.T("Saving "),
			printer.TS("server", printer.ComponentStyle),
			printer.T(" certificate to file: "),
			printer.TS(certFile.Name(), printer.FilePathStyle),
		)
	} else {
		certStyledTexts = append(
			certStyledTexts,
			printer.T("Adding "),
			printer.TS("server", printer.ComponentStyle),
			printer.T(" certificate to "),
			printer.TS(store, printer.LiteralStyle),
			printer.T(" store"),
		)
		if store == "user" {
			certStyledTexts = append(
				certStyledTexts,
				printer.T(", accept the dialog"),
			)
		} else {
			if _, _, err := launcherCommon.ConfigAdminAgent(false); err != nil {
				certStyledTexts = append(
					certStyledTexts,
					printer.T(", authorize "),
					printer.TS("config-admin-agent", printer.ComponentStyle),
					printer.T(" if needed"),
				)
			}
		}
		if store == "local" {
			addLocalCertData = serverCertificate.Raw
		} else {
			addUserCertData = serverCertificate.Raw
		}
	}
	certStyledTexts = append(certStyledTexts, printer.T("..."))
	fmt.Print(printer.Gen(printer.Configuration, "", certStyledTexts...))
	if result := executor.RunSetUp(&executor.RunSetUpOptions{AddUserCertData: addUserCertData, AddLocalCertData: addLocalCertData, CertFilePath: c.CertFilePath()}); !result.Success() {
		printer.PrintFailedResultError(result)
		errorCode = internal.ErrConfigCertAdd
		return
	} else {
		printer.PrintSucceeded()
	}
	if store != "tmp" {
		for _, host := range hosts {
			if !server.CheckConnectionFromServer(host, false) {
				printer.Println(
					printer.Error,
					printer.TS(host, printer.LiteralStyle),
					printer.T(" must have been trusted automatically at this point."),
				)
				errorCode = internal.ErrServerConnectSecure
				return
			} else if !server.LanServerHost(serverId, c.gameTitle, host, false) {
				printer.Println(
					printer.Error,
					printer.T("Something went wrong, "),
					printer.TS(host, printer.LiteralStyle),
					printer.T(" does not point to the LAN "),
					printer.TS("server", printer.ComponentStyle),
					printer.T(" or there is a certificate issue."),
				)
				errorCode = internal.ErrTrustCert
				return
			}
		}
	}
	return
}
