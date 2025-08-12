package cmd

import (
	"github.com/spf13/cobra"
)

var RemoveLocalCert bool
var UnmapIP bool
var UnmapCDN bool
var RemoveAll bool

func InitRevert(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&UnmapIP,
		"ip",
		"i",
		false,
		"Remove the IP mappings from the local DNS server",
	)
	cmd.Flags().BoolVarP(
		&UnmapCDN,
		"CDN",
		"c",
		false,
		"Remove the CDN mappings from the local DNS server",
	)
	cmd.Flags().BoolVarP(
		&RemoveLocalCert,
		"localCert",
		"l",
		false,
		"Remove the certificate from the local machine's trusted root store",
	)
	cmd.Flags().BoolVarP(
		&RemoveAll,
		"all",
		"a",
		false,
		"Removes all configuration. Equivalent to the rest of the flags being set without fail-fast.",
	)
}
