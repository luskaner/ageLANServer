//go:build !go1.20
// +build !go1.20

package internal

import "net"

func FlagsExtraCheck(_ net.Flags) bool {
	return true
}
