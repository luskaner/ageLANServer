package cmd

import (
	"os"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server-genCert/internal"
)

var (
	osExecutableFn            = os.Executable
	certificatePairFolderFn   = common.CertificatePairFolder
	certificatePairsFn        = common.CertificatePairs
	generateCertificatePairsFn = internal.GenerateCertificatePairs
)
