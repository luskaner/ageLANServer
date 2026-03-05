package cmd

import (
	"encoding/base64"
	"net"

	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

var MapIP net.IP
var AddLocalCertData []byte
var AddLocalCertDataB64 string
var GameId string

func InitSetUp(flags *pflag.FlagSet) {
	flags.IPVarP(
		&MapIP,
		"ip",
		"i",
		nil,
		"IP to resolve in local DNS server.",
	)
	flags.StringVarP(
		&AddLocalCertDataB64,
		"localCert",
		"l",
		"",
		"Add the certificate to the local machine's trusted root store",
	)
	commonCmd.GameVarCommand(flags, &GameId)
}

func DecodeSetUpFlags() error {
	if AddLocalCertDataB64 != "" {
		b, err := base64.StdEncoding.DecodeString(AddLocalCertDataB64)
		if err != nil {
			return err
		}
		AddLocalCertData = b
	}
	return nil
}
