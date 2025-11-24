package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/server-genCert/internal"
	"github.com/spf13/cobra"
)

var replace bool
var Version string

var (
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "genCert generates a CA and certificate",
		Run: func(_ *cobra.Command, _ []string) {
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
			} else {
				fmt.Println("Certificate pair generated successfully.")
			}
		},
	}
)

func Execute() error {
	rootCmd.Version = Version
	rootCmd.Flags().BoolVarP(&replace, "replace", "r", false, "Overwrite existing certificate pair.")
	return rootCmd.Execute()
}
