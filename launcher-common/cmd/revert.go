package cmd

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/spf13/cobra"
)

var RemoveLocalCert bool
var UnmapIPs bool
var UnmapCDN bool
var RemoveAll bool

func InitRevert(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&UnmapIPs,
		"ip",
		"i",
		false,
		fmt.Sprintf("Remove the IP mappings from the local DNS server that resolve to '%s'", common.Domain),
	)
	cmd.Flags().BoolVarP(
		&UnmapCDN,
		"CDN",
		"c",
		false,
		fmt.Sprintf("Remove the mappings from the local DNS server that resolve %s to '%s'", launcherCommon.CDNIP, common.Domain),
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
