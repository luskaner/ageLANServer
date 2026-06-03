package hosts

import "golang.org/x/net/idna"

type Host string

func (h Host) String() string {
	host, _ := idna.Lookup.ToASCII(string(h))
	return host
}
