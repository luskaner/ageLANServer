package cmd

import (
	"common"
	"fmt"
	"github.com/spf13/cobra"
	"net"
)

var MapIPs []net.IP
var AddLocalCertData []byte

func InitSetUp(cmd *cobra.Command) {
	cmd.Flags().IPSliceVarP(
		&MapIPs,
		"ip",
		"i",
		nil,
		fmt.Sprintf("IP to resolve to '%s' in local DNS server (up to 9)", common.Domain),
	)
	cmd.Flags().BytesBase64VarP(
		&AddLocalCertData,
		"localCert",
		"l",
		nil,
		"Add the certificate to the local machine's trusted root store",
	)
}
