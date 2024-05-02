package login

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type request struct {
	AccountType    string `form:"accountType"`
	PlatformUserId uint64 `form:"platformUserID"`
	Alias          string `form:"alias"`
}

func Platformlogin(c *gin.Context) {
	t := time.Now().UTC().Unix()
	var req request
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, i.A{2, "", 0, t, i.A{}, i.A{}, 0, 0, nil, nil, i.A{}, i.A{}, 0, i.A{}})
		return
	}
	t2 := t - rand.Int63n(3600*2-3600+1) + 3600
	t3 := t - rand.Int63n(3600*2-3600+1) + 3600
	u := models.GetOrCreateUser(req.AccountType == "XBOXLIVE", req.PlatformUserId, req.Alias)
	u.SetPresence(1)
	sess, ok := models.GetSessionByUser(u)
	if ok {
		sess.Delete()
	}
	sessionId := models.CreateSession(u)
	profileInfo := u.GetProfileInfo(false)
	var config []i.A
	var rawConfig = files.KeyedFiles["configuration.json"]
	for _, key := range rawConfig.Keys() {
		value, _ := rawConfig.Get(key)
		config = append(config, i.A{key, value})
	}
	profileId := u.GetProfileId()
	response := i.A{
		0,
		sessionId,
		549_000_000,
		t,
		i.A{
			profileId,
			u.GetPlatformPath(),
			u.GetPlatformId(),
			-1,
			0,
			"en",
			"eur",
			2,
			nil,
		},
		i.A{profileInfo},
		0,
		0,
		nil,
		config,
		i.A{
			0,
			profileInfo,
			i.A{
				0,
				i.A{},
				i.A{},
				i.A{},
				i.A{},
				i.A{},
				i.A{},
				i.A{},
			},
			i.A{u.GetExtraProfileInfo()},
			i.A{
				i.A{2, profileId, 0, "", t2},
				i.A{39, profileId, 671, "", t2},
				i.A{41, profileId, 191, "", t2},
				i.A{42, profileId, 480, "", t2},
				i.A{44, profileId, 0, "", t2},
				i.A{45, profileId, 0, "", t2},
				i.A{46, profileId, 0, "", t2},
				i.A{47, profileId, 0, "", t2},
				i.A{48, profileId, 0, "", t2},
				i.A{50, profileId, 0, "", t2},
				i.A{60, profileId, 1, "", t2},
				i.A{142, profileId, 1, "", t3},
				i.A{171, profileId, 1, "", t2},
				i.A{172, profileId, 4, "", t2},
				i.A{173, profileId, 1, "", t2},
			},
			nil,
			i.A{},
			nil,
			1,
			i.A{},
		},
		i.A{},
		0,
		i.A{i.A{"", nil, "127.0.0.1", 27012, 27112, 27212}},
	}
	expiration := time.Now().Add(time.Hour).UTC().Format(time.RFC1123)
	c.Header("Set-Cookie", fmt.Sprintf("reliclink=%d; Expires=%s; Max-Age=3600", u.GetReliclink(), expiration))
	c.Header("Request-Context", "appId=cid-v1:d21b644d-4116-48ea-a602-d6167fb46535")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(http.StatusOK, response)
}
