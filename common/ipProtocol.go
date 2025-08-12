package common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
)

type IPProtocol string

const (
	IPvDual IPProtocol = ""
	IPv4    IPProtocol = "v4"
	IPv6    IPProtocol = "v6"
	IPv4v6  IPProtocol = "v4+v6"
)

var supportedIPProtocols = mapset.NewThreadUnsafeSet(IPvDual, IPv4, IPv6, IPv4v6)

func (i *IPProtocol) IPv4() bool {
	return *i != IPv6
}

func (i *IPProtocol) IPv6() bool {
	return *i != IPv4
}

func (i *IPProtocol) Set(val string) error {
	ipProtocol := IPProtocol(val)
	if !supportedIPProtocols.Contains(ipProtocol) {
		return fmt.Errorf("%v is not a supported IPProtocol", val)
	}
	*i = ipProtocol
	return nil
}

func (i *IPProtocol) Type() string {
	return "IPProtocol"
}

func (i *IPProtocol) String() string {
	return string(*i)
}

func (i *IPProtocol) UnmarshalText(text []byte) error {
	return i.Set(string(text))
}
