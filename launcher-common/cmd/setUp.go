package cmd

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/spf13/cobra"
	"net"
)

var MapIPs []net.IP
var AddLocalCertData []byte
var MapCDN bool

func InitSetUp(cmd *cobra.Command) {
	cmd.Flags().IPSliceVarP(
		&MapIPs,
		"ip",
		"i",
		nil,
		fmt.Sprintf("IP to resolve to '%s' in local DNS server (up to 9)", common.Domain),
	)
	cmd.Flags().BoolVarP(
		&MapCDN,
		"CDN",
		"c",
		false,
		fmt.Sprintf("Resolve '%s' to %s in local DNS server", launcherCommon.CDNIP, launcherCommon.CDNDomain),
	)
	cmd.Flags().BytesBase64VarP(
		&AddLocalCertData,
		"localCert",
		"l",
		nil,
		"Add the certificate to the local machine's trusted root store",
	)
}
