# Server GenCert
[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/aoe2DELanServer/server-genCert)](https://goreportcard.com/report/github.com/luskaner/aoe2DELanServer/server-genCert)

Generates the SSL certificate needed for server. Generates the `resources/certificates/key.pem`
and `resources/certificates/cert.pem` upon execution. Resides in `bin` subdirectory.

## Command Line

CLI is available and you can see the available options with
`genCert -h`.

## Exit Codes

* [Base codes](../common/errors.go).
* [Server genCert codes](internal/errors.go).
