package router

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/luskaner/ageLANServer/common"
)

type Proxy struct {
	Router
	proxy            *httputil.ReverseProxy
	host             string
	initializeRoutes func(gameId string, next http.Handler) http.Handler
}

func NewProxy(host string, initFn func(gameId string, next http.Handler) http.Handler) (proxy Proxy) {
	proxy = Proxy{host: host, initializeRoutes: initFn}
	ip, err := common.DirectHostToIP(host)
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
