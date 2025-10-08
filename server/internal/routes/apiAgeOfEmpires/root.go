package apiAgeOfEmpires

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models/apiAgeOfEmpires"
	"github.com/miekg/dns"
)

func resolveDirect(host string, dnsServer string) (string, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)

	c := new(dns.Client)
	in, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		return "", err
	}

	for _, ans := range in.Answer {
		if a, ok := ans.(*dns.A); ok {
			return a.A.String(), nil
		}
	}
	return "", fmt.Errorf("no IP found for %s", host)
}

func Root() *httputil.ReverseProxy {
	ip, err := resolveDirect(common.ApiAgeOfEmpires, "8.8.8.8:53")
	if err != nil {
		return nil
	}
	remote, err := url.Parse(fmt.Sprintf("https://%s", ip))
	if err != nil {
		return nil
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: common.ApiAgeOfEmpires,
		},
	}
	proxy.Director = func(req *http.Request) {
		req.URL.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Path = strings.TrimPrefix(req.URL.Path, apiAgeOfEmpires.Prefix)
	}
	return proxy
}
