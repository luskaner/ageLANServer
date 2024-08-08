package login

import (
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/files"
	models2 "github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
	"time"
)

type request struct {
	AccountType    string `schema:"accountType"`
	PlatformUserId uint64 `schema:"platformUserID"`
	Alias          string `schema:"alias"`
}

func Platformlogin(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC().Unix()
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2, "", 0, t, i.A{}, i.A{}, 0, 0, nil, nil, i.A{}, i.A{}, 0, i.A{}})
		return
	}
	t2 := t - i.Rng.Int63n(3600*2-3600+1) + 3600
	t3 := t - i.Rng.Int63n(3600*2-3600+1) + 3600
	u := models2.GetOrCreateUser(r.RemoteAddr, req.AccountType == "XBOXLIVE", req.PlatformUserId, req.Alias)
	u.SetPresence(1)
	sess, ok := models2.GetSessionByUser(u)
	if ok {
		sess.Delete()
	}
	sessionId := models2.CreateSession(u)
	profileInfo := u.GetProfileInfo(false)
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
		files.Login,
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
	w.Header().Set("Set-Cookie", fmt.Sprintf("reliclink=%d; Expires=%s; Max-Age=3600", u.GetReliclink(), expiration))
	w.Header().Set("Request-Context", "appId=cid-v1:d21b644d-4116-48ea-a602-d6167fb46535")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	i.JSON(&w, response)
}
