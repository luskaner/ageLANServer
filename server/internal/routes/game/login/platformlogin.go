package login

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/relationship"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

type request struct {
	AccountType      string `schema:"accountType"`
	PlatformUserId   uint64 `schema:"platformUserID"`
	Alias            string `schema:"alias"`
	GameId           string `schema:"title"`
	MacAddress       string `schema:"macAddress"`
	ClientLibVersion uint16 `schema:"clientLibVersion"`
}

func Platformlogin(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC().Unix()
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2, "", 0, t, i.A{}, i.A{}, 0, 0, nil, nil, i.A{}, i.A{}, 0, i.A{}})
		return
	}
	var t2 int64
	var t3 int64
	i.WithRng(func(rand *rand.Rand) {
		t2 = t - rand.Int64N(3600*2-3600+1) + 3600
		t3 = t - rand.Int64N(3600*2-3600+1) + 3600
	})
	game := models.G(r)
	title := game.Title()
	users := game.Users()
	u := users.GetOrCreateUser(title, r.RemoteAddr, req.MacAddress, req.AccountType == "XBOXLIVE", req.PlatformUserId, req.Alias)
	sess, ok := models.GetSessionByUserId(u.GetId())
	if ok {
		sess.Delete()
	}
	sessionId := models.CreateSession(req.GameId, u.GetId(), req.ClientLibVersion)
	sess, _ = models.GetSessionById(sessionId)
	relationship.ChangePresence(req.GameId, req.ClientLibVersion, users, u, 1)
	profileInfo := u.GetProfileInfo(false, req.GameId, req.ClientLibVersion)
	if title == common.GameAoE3 || title == common.GameAoM {
		for user := range users.GetUserIds() {
			if user != u.GetId() {
				currentSess, currentOk := models.GetSessionByUserId(user)
				if currentOk {
					wss.SendOrStoreMessage(
						currentSess,
						"FriendAcceptMessage",
						i.A{profileInfo},
					)
				}
			}
		}
	}
	profileId := u.GetProfileId()
	extraProfileInfoList := i.A{}
	if title == common.GameAoE2 {
		extraProfileInfoList = append(extraProfileInfoList, u.GetExtraProfileInfo(req.GameId, req.ClientLibVersion))
	}
	var unknownProfileInfoList i.A
	switch title {
	case common.GameAoE2:
		unknownProfileInfoList = i.A{
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
		}
	case common.GameAoE3, common.GameAoM:
		unknownProfileInfoList = i.A{
			i.A{291, u.GetId(), 16, "", t2},
		}
	default:
		unknownProfileInfoList = i.A{}
	}
	battleServers := game.BattleServers()
	servers := battleServers.Encode(r)
	if len(servers) == 0 {
		server := battleServers.NewBattleServer("")
		server.IPv4 = "127.0.0.1"
		server.BsPort = 27012
		server.WebSocketPort = 27112
		if title != common.GameAoE1 {
			server.Name = "localhost"
			server.OutOfBandPort = 27212
		}
		servers = append(servers, server.EncodeLogin(r))
	}
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
	}
	if title == common.GameAoE1 {
		response = append(response, i.A{})
	}
	allProfileInfo := i.A{
		0,
		profileInfo,
		relationship.Relationships(title, req.ClientLibVersion, users, u),
		extraProfileInfoList,
		unknownProfileInfoList,
		nil,
		i.A{},
		nil,
		1,
	}
	if title != common.GameAoE1 {
		allProfileInfo = append(allProfileInfo, i.A{})
	}
	response = append(response,
		game.Resources().LoginData,
		allProfileInfo,
		i.A{},
		0,
		servers,
	)
	expiration := time.Now().Add(time.Hour).UTC().Format(time.RFC1123)
	w.Header().Set("Set-Cookie", fmt.Sprintf("reliclink=%d; Expires=%s; Max-Age=3600", u.GetReliclink(), expiration))
	w.Header().Set("Request-Context", "appId=cid-v1:d21b644d-4116-48ea-a602-d6167fb46535")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	i.JSON(&w, response)
}
