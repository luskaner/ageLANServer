package client

import (
	"crypto/tls"
)

var TlsClientConfig *tls.Config

func init() {
	TlsClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
}
