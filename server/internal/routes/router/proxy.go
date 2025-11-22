package router

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/miekg/dns"
)

// Google, Cloudfare and OpenDNS primaries then secondaries
var dnsServers = []string{"8.8.8.8", "1.1.1.1", "208.67.222.222", "8.8.4.4", "1.0.0.1", "208.67.220.220"}

func resolveDirect(host string) (string, error) {
	for _, dnsServer := range dnsServers {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(host), dns.TypeA)

		c := new(dns.Client)
		in, _, err := c.Exchange(m, dnsServer+":53")
		if err != nil {
			continue
		}

		for _, ans := range in.Answer {
			if a, ok := ans.(*dns.A); ok {
				return a.A.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IP found for %s", host)
}

type Proxy struct {
	Router
	proxy            *httputil.ReverseProxy
	host             string
	initializeRoutes func(gameId string, next http.Handler) http.Handler
}

func NewProxy(host string, initFn func(gameId string, next http.Handler) http.Handler) (proxy Proxy) {
	proxy = Proxy{host: host, initializeRoutes: initFn}
	ip, err := resolveDirect(host)
	if err != nil {
		return
	}
	remote, err := url.Parse(fmt.Sprintf("https://%s", ip))
	if err != nil {
		return
	}
	proxy.proxy = httputil.NewSingleHostReverseProxy(remote)
	proxy.proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: host,
		},
	}
	proxy.proxy.Director = func(req *http.Request) {
		req.URL.Host = remote.Host
		req.URL.Scheme = remote.Scheme
	}
	return
}

func (p *Proxy) Name() string {
	return fmt.Sprintf("proxy %s", p.host)
}

func (p *Proxy) Check(r *http.Request) bool {
	host := r.Host
	if newHost, _, err := net.SplitHostPort(r.Host); err == nil {
		host = newHost
	}
	return strings.ToLower(host) == strings.ToLower(p.host)
}

func (p *Proxy) InitializeRoutes(gameId string, next http.Handler) http.Handler {
	p.initialize()
	current := p.initializeRoutes(gameId, next)
	if p.proxy != nil {
		p.group.HandlePath("/", p.proxy)
	}
	return current
}
