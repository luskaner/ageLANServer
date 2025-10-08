package cmdUtils

import (
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
)

func ResolveSSLFilesPath(gameId string, auto bool) (resolvedCertFile string, resolvedKeyFile string, err error) {
	if auto {
		fmt.Println("Auto resolving SSL certificate and key files...")
		serverExe := common.FindExecutablePath(common.GetExeFileName(true, common.Server))
		if serverExe == "" {
			err = fmt.Errorf("could not find server executable")
			return
		}
		ok, _, cert, key, _, selfSignedCert, selfSignedKey := common.CertificatePairs(serverExe)
		if !ok {
			err = fmt.Errorf("no SSL certificate and keys found")
		}
		if gameId == common.GameAoM {
			resolvedCertFile = cert
			resolvedKeyFile = key
		} else {
			resolvedCertFile = selfSignedCert
			resolvedKeyFile = selfSignedKey
		}
		return
	}
	var f os.FileInfo
	var path string
	if f, path, err = common.ParsePath("SSL.CertFile", nil); err != nil || f.IsDir() {
		err = fmt.Errorf("invalid certificate file")
		return
	} else {
		resolvedCertFile = path
	}
	if f, path, err = common.ParsePath("SSL.KeyFile", nil); err != nil || f.IsDir() {
		err = fmt.Errorf("invalid key file")
		return
	} else {
		resolvedKeyFile = path
	}
	return
}
