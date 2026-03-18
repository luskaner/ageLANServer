package hosts

import "golang.org/x/net/idna"

type Host string

func (h Host) String() string {
	norm, _ := idna.Punycode.ToASCII(string(h))
	return norm
}
