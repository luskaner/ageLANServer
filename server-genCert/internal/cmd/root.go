package cmd

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server-genCert/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
)

var replace bool
var Version string

var (
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "genCert generates a self-signed certificate",
		Run: func(_ *cobra.Command, _ []string) {
			exe, err := os.Executable()
			if err != nil {
				fmt.Println("Could not get executable path")
				os.Exit(common.ErrGeneral)
			}
			serverExe := path.Join(filepath.Dir(filepath.Dir(exe)), common.GetExeFileName(true, common.Server))
			serverFolder := common.CertificatePairFolder(serverExe)
			if serverFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				os.Exit(internal.ErrCertDirectory)
			}
			if !replace {
				if exists, _, _ := common.CertificatePair(serverExe); exists {
					fmt.Println("Already have certificate pair and force is false, set force to true or delete it manually.")
					os.Exit(internal.ErrCertCreateExisting)
				}
			}
			if !internal.GenerateCertificatePair(serverFolder) {
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
	rootCmd.PersistentFlags().BoolVarP(&replace, "replace", "r", false, "Overwrite existing certificate pair.")
	return rootCmd.Execute()
}
