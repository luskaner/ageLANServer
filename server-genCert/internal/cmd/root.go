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

func rootCmd(_ *pflag.FlagSet) error {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Could not get executable path")
		os.Exit(common.ErrGeneral)
	}
	serverExe := filepath.Join(filepath.Dir(filepath.Dir(exe)), executables.Filename(true, executables.Server))
	serverFolder := common.CertificatePairFolder(serverExe)
	if serverFolder == "" {
		fmt.Println("Failed to determine certificate pairs folder")
		os.Exit(internal.ErrCertDirectory)
	}
	if !replace {
		if exists, _, _, _, _, _, _ := common.CertificatePairs(serverExe); exists {
			fmt.Println("Already have certificate pairs and force is false, set force to true or delete it manually.")
			os.Exit(internal.ErrCertCreateExisting)
		}
	}
	if !internal.GenerateCertificatePairs(serverFolder) {
		fmt.Println("Could not generate certificate pair.")
		os.Exit(internal.ErrCertCreate)
	}

	fmt.Println("Certificate pair generated successfully.")
	return nil
}

func Execute() error {
	singleFlagSet := cmd.NewSingleFlagSet(rootCmd, Version)
	singleFlagSet.Fs().BoolVarP(&replace, "replace", "r", false, "Overwrite existing certificate pair.")
	return singleFlagSet.Execute()
}
