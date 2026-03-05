package cmd

import (
	"github.com/spf13/pflag"
)

var RemoveLocalCert bool
var UnmapIPs bool
var RemoveAll bool

func InitRevert(flags *pflag.FlagSet) {
	flags.BoolVarP(
		&UnmapIPs,
		"ip",
		"i",
		false,
		"Remove the IP mappings from the local DNS server",
	)
	flags.BoolVarP(
		&RemoveLocalCert,
		"localCert",
		"l",
		false,
		"Remove the certificate from the local machine's trusted root store",
	)
	flags.BoolVarP(
		&RemoveAll,
		"all",
		"a",
		false,
		"Removes all configuration. Equivalent to the rest of the flags being set without fail-fast.",
	)
}
