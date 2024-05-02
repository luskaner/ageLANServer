package cloud

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/j"
	"aoe2DELanServer/routes/game/cloud/extra"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
	"time"
)

func GetTempCredentials(c *gin.Context) {
	fullKey := c.Query("key")
	key := strings.TrimPrefix(fullKey, "/cloudfiles/")
	info := extra.Create(key)
	t := info.GetExpiry()
	tUnix := t.Unix()
	for _, file := range asset.CloudFiles {
		if file.Key == key {
			se := url.QueryEscape(t.Format(time.RFC3339))
			sv := url.QueryEscape(file.Version)
			c.JSON(200, j.A{0, tUnix, fmt.Sprintf("sig=%s&se=%s&sv=%s&sp=r&sr=b", url.QueryEscape(info.GetSignature()), se, sv), fullKey})
			return
		}
	}
	c.JSON(200, j.A{2, t, "", fullKey})
}
