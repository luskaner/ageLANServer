package login

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/j"
	"aoe2DELanServer/session"
	"aoe2DELanServer/user"
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
		c.JSON(http.StatusOK, j.A{2, "", 0, t, j.A{}, j.A{}, 0, 0, nil, nil, j.A{}, j.A{}, 0, j.A{}})
		return
	}
	t2 := t - rand.Int63n(3600*2-3600+1) + 3600
	t3 := t - rand.Int63n(3600*2-3600+1) + 3600
	u := user.GetOrCreate(req.AccountType == "XBOXLIVE", req.PlatformUserId, req.Alias)
	u.SetPresence(1)
	sess, ok := session.GetByUser(u)
	if ok {
		sess.Delete()
	}
	sessionId := session.Create(u)
	profileInfo := u.GetProfileInfo(false)
	var config []j.A
	var rawConfig = asset.KeyedFiles["configuration.json"]
	for _, key := range rawConfig.Keys() {
		value, _ := rawConfig.Get(key)
		config = append(config, j.A{key, value})
	}
	profileId := u.GetProfileId()
	response := j.A{
		0,
		sessionId,
		549_000_000,
		t,
		j.A{
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
		j.A{profileInfo},
		0,
		0,
		nil,
		config,
		j.A{
			0,
			profileInfo,
			j.A{
				0,
				j.A{},
				j.A{},
				j.A{},
				j.A{},
				j.A{},
				j.A{},
				j.A{},
			},
			j.A{u.GetExtraProfileInfo()},
			j.A{
				j.A{2, profileId, 0, "", t2},
				j.A{39, profileId, 671, "", t2},
				j.A{41, profileId, 191, "", t2},
				j.A{42, profileId, 480, "", t2},
				j.A{44, profileId, 0, "", t2},
				j.A{45, profileId, 0, "", t2},
				j.A{46, profileId, 0, "", t2},
				j.A{47, profileId, 0, "", t2},
				j.A{48, profileId, 0, "", t2},
				j.A{50, profileId, 0, "", t2},
				j.A{60, profileId, 1, "", t2},
				j.A{142, profileId, 1, "", t3},
				j.A{171, profileId, 1, "", t2},
				j.A{172, profileId, 4, "", t2},
				j.A{173, profileId, 1, "", t2},
			},
			nil,
			j.A{},
			nil,
			1,
			j.A{},
		},
		j.A{},
		0,
		j.A{j.A{"", nil, "127.0.0.1", 27012, 27112, 27212}},
	}
	expiration := time.Now().Add(time.Hour).UTC().Format(time.RFC1123)
	c.Header("Set-Cookie", fmt.Sprintf("reliclink=%d; Expires=%s; Max-Age=3600", u.GetReliclink(), expiration))
	c.Header("Request-Context", "appId=cid-v1:d21b644d-4116-48ea-a602-d6167fb46535")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.JSON(http.StatusOK, response)
}
