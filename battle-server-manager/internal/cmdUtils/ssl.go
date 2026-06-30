package cmdUtils

import (
	"battle-server-manager/internal"
	"fmt"
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
)

func ResolveSSLFilesPath(gameId string, ssl internal.SSL) (resolvedCertFile string, resolvedKeyFile string, err error) {
	if ssl.Auto {
		commonLogger.Println("Auto resolving SSL certificate and key files...")
		serverExe := executables.FindPath(executables.NativeFileName(true, executables.Server))
		if serverExe == "" {
			err = fmt.Errorf("could not find server executable")
			return
		}
		ok, _, cert, key, _, selfSignedCert, selfSignedKey := common.CertificatePairs(serverExe)
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
	var f os.FileInfo
	var path string
	if f, path, err = common.ParsePath(common.EnhancedViperStringToStringSlice(ssl.CertFile), nil); err != nil || f.IsDir() {
		err = fmt.Errorf("invalid certificate file")
		return
	}
	resolvedCertFile = path
	if f, path, err = common.ParsePath(common.EnhancedViperStringToStringSlice(ssl.KeyFile), nil); err != nil || f.IsDir() {
		err = fmt.Errorf("invalid key file")
		return
	}
	resolvedKeyFile = path
	return
}
