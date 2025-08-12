package cmd

import (
	"fmt"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/spf13/cobra"
)

var MapIPAddrValue NetIPAddrValue
var AddLocalCertData []byte
var MapCDN bool

func InitSetUp(cmd *cobra.Command) {
	cmd.Flags().VarP(
		&MapIPAddrValue,
		"ip",
		"i",
		"IP to resolve in local DNS server",
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
