package cmd

import (
	"net"

	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/cobra"
)

var MapIP net.IP
var AddLocalCertData []byte
var GameId string

func InitSetUp(cmd *cobra.Command) {
	cmd.Flags().IPVarP(
		&MapIP,
		"ip",
		"i",
		nil,
		"IP to resolve in local DNS server.",
	)
	cmd.Flags().BytesBase64VarP(
		&AddLocalCertData,
		"localCert",
		"l",
		nil,
		"Add the certificate to the local machine's trusted root store",
	)
	commonCmd.GameVarCommand(cmd.Flags(), &GameId)
	err := cmd.MarkFlagRequired("game")
	if err != nil {
		panic(err)
	}
}
