package cmd

import (
	"fmt"
	"net/netip"
)

type NetIPAddrValue struct {
	Addr netip.Addr
}

func (a *NetIPAddrValue) String() string {
	return a.Addr.String()
}

func (a *NetIPAddrValue) Set(s string) error {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return err
	}
	if addr.Zone() != "" {
		return fmt.Errorf("IP cannot have a zone even if it's IPv6")
	}
	a.Addr = addr
	return nil
}

func (a *NetIPAddrValue) Type() string {
	return "netip.Addr"
}
