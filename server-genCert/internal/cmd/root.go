package cmd

import (
	"common"
	"fmt"
	"genCert/internal"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var replace bool
var Version string

var (
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "genCert generates a self-signed certificate to act as " + common.Domain,
		Run: func(_ *cobra.Command, _ []string) {
			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				os.Exit(internal.ErrCertDirectory)
			}
			if !replace && common.HasCertificatePair(os.Args[0]) {
				fmt.Println("Already have certificate pair and force is false, set force to true or delete it manually.")
				os.Exit(internal.ErrCertCreateExisting)
			}
			if !internal.GenerateCertificatePair(certificatePairFolder) {
				fmt.Println("Could not generate certificate pair.")
				os.Exit(internal.ErrCertCreate)
			} else {
				fmt.Println("Certificate pair generated successfully.")
			}
		},
	}
)

func Execute() error {
	rootCmd.PersistentFlags().BoolVarP(&replace, "replace", "r", false, "Overwrite existing certificate pair.")
	rootCmd.PersistentFlags().StringVarP(&Version, "version", "v", Version, "Version")
	return rootCmd.Execute()
}
