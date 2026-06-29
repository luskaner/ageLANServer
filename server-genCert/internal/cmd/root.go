package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/server-genCert/internal"
	"github.com/spf13/pflag"
)

var replace bool
var Version string

func rootCmd(_ *pflag.FlagSet) (err error, exitCode int) {
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
	if !replace {
		if exists, _, _, _, _, _, _ := common.CertificatePairs(serverExe); exists {
			fmt.Println("Already have certificate pairs and force is false, set force to true or delete it manually.")
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
	singleFlagSet := cmd.NewSingleFlagSet(rootCmd, Version)
	singleFlagSet.Fs().BoolVarP(&replace, "replace", "r", false, "Overwrite existing certificate pair.")
	return singleFlagSet.Execute()
}
