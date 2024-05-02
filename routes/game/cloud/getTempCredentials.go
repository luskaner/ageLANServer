package cloud

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
	"time"
)

func GetTempCredentials(c *gin.Context) {
	fullKey := c.Query("key")
	key := strings.TrimPrefix(fullKey, "/cloudfiles/")
	info := models.CreateCredentials(key)
	t := info.GetExpiry()
	tUnix := t.Unix()
	for _, file := range files.CloudFiles {
		if file.Key == key {
			se := url.QueryEscape(t.Format(time.RFC3339))
			sv := url.QueryEscape(file.Version)
			c.JSON(200, i.A{0, tUnix, fmt.Sprintf("sig=%s&se=%s&sv=%s&sp=r&sr=b", url.QueryEscape(info.GetSignature()), se, sv), fullKey})
			return
		}
	}
	c.JSON(200, i.A{2, t, "", fullKey})
}
