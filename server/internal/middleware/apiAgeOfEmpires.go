package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models/apiAgeOfEmpires"
)

func isApiAgeOfEmpires(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, apiAgeOfEmpires.Prefix)
}

func ApiAgeOfEmpiresMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subdomain, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.ApiAgeOfEmpiresDomain && subdomain == common.ApiAgeOfEmpiresSubdomain {
			r.URL.Path = path.Join(apiAgeOfEmpires.Prefix, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
