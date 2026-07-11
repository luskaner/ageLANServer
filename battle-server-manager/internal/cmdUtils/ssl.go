package cmdUtils

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
)

func keyCert(parentFolder, gameId string) (resolvedCertFile string, resolvedKeyFile string, err error) {
	ok, cert, key, _, selfSignedCert, selfSignedKey := common.CertificatePairs(parentFolder)
	if !ok {
		err = fmt.Errorf("no SSL certificate and keys found")
	}
	if gameId == game.AoE4 || gameId == game.AoM {
		resolvedCertFile = cert
		resolvedKeyFile = key
	} else {
		resolvedCertFile = selfSignedCert
		resolvedKeyFile = selfSignedKey
	}
	return
}

func ResolveSSLFilesPath(gameId string, certsPath string) (resolvedCertFile string, resolvedKeyFile string, err error) {
	if certsPath == "auto" {
		commonLogger.Println("Auto resolving SSL certificate and key files...")
		serverExe := executables.FindPath(executables.NativeFileName(true, executables.Server))
		if serverExe == "" {
			err = fmt.Errorf("could not find server executable")
			return
		}
		certsPath = common.CertificatePairFolder(serverExe)
	}
	return keyCert(certsPath, gameId)
}
