package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/cmd/genCert"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/server-genCert/internal"
	"github.com/spf13/pflag"
)

var (
	Version string
	values  *genCert.Values
)

func runRoot(_ *pflag.FlagSet) (err error, exitCode int) {
	var exe string
	exe, err = os.Executable()
	if err != nil {
		fmt.Println("Could not get executable path")
		exitCode = common.ErrGeneral
		return
	}
	serverExe := filepath.Join(filepath.Dir(filepath.Dir(exe)), executables.NativeFileName(true, executables.Server))
	serverFolder := common.CertificatePairFolder(serverExe)
	if serverFolder == "" {
		fmt.Println("Failed to determine certificate pairs folder")
		exitCode = internal.ErrCertDirectory
		return
	}
	if !values.Replace {
		certificateFolder := common.CertificatePairFolder(serverExe)
		if exists, _, _, _, _, _ := common.CertificatePairs(certificateFolder); exists {
			fmt.Println("Already have certificate pairs and replace is false, set replace to true or delete it manually.")
			if values.IgnoreIfExisting {
				return
			}
			exitCode = internal.ErrCertCreateExisting
			return
		}
	}
	if !internal.GenerateCertificatePairs(serverFolder) {
		fmt.Println("Could not generate certificate pair.")
		exitCode = internal.ErrCertCreate
		return
	}
	fmt.Println("Certificate pair generated successfully.")
	return
}

func Execute() (err error, exitCode int) {
	var singleFs *cmd.SingleFlagSet
	values, singleFs = genCert.SingleFlagSet(Version, runRoot)
	return singleFs.Execute()
}
