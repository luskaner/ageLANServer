package cloud

import (
	"fmt"
	"net/http"
	"net/url"
	"server/files"
	i "server/internal"
	"server/models"
	"strings"
	"time"
)

func GetTempCredentials(w http.ResponseWriter, r *http.Request) {
	fullKey := r.URL.Query().Get("key")
	key := strings.TrimPrefix(fullKey, "/cloudfiles/")
	info := models.CreateCredentials(key)
	t := info.GetExpiry()
	tUnix := t.Unix()
	for _, file := range files.CloudFiles {
		if file.Key == key {
			se := url.QueryEscape(t.Format(time.RFC3339))
			sv := url.QueryEscape(file.Version)
			i.JSON(&w, i.A{0, tUnix, fmt.Sprintf("sig=%s&se=%s&sv=%s&sp=r&sr=b", url.QueryEscape(info.GetSignature()), se, sv), fullKey})
			return
		}
	}
	i.JSON(&w, i.A{2, t, "", fullKey})
}
