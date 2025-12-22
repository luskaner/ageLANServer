package login

import (
	"fmt"
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
	game := models.G(r)
	title := game.Title()
	users := game.Users()
	sessions := game.Sessions()
	var avatarStatDefinitions models.AvatarStatDefinitions = nil
	if title != common.GameAoE1 {
		avatarStatDefinitions = game.LeaderboardDefinitions().AvatarStatDefinitions()
	}
	u := users.GetOrCreateUser(
		title,
		avatarStatDefinitions,
		r.RemoteAddr,
		req.MacAddress,
		req.AccountType == "XBOXLIVE",
		req.PlatformUserId,
		req.Alias,
	)
	sess, ok := sessions.GetByUserId(u.GetId())
	if ok {
		sessions.Delete(sess.Id())
	}
	sessionId := sessions.Create(u.GetId(), req.ClientLibVersion)
	sess, _ = sessions.GetById(sessionId)
	relationship.ChangePresence(req.ClientLibVersion, sessions, users, u, 1)
	profileInfo := u.GetProfileInfo(false, req.ClientLibVersion)
	if title == common.GameAoE3 || title == common.GameAoM {
		for user := range users.GetUserIds() {
			if user != u.GetId() {
				currentSess, currentOk := sessions.GetByUserId(user)
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
		extraProfileInfoList = append(extraProfileInfoList, u.GetExtraProfileInfo(req.ClientLibVersion))
	}
	battleServers := game.BattleServers()
	servers := battleServers.Encode(r)
	if len(servers) == 0 {
		server := battleServers.NewBattleServer("")
		server.SetIPv4("127.0.0.1")
		server.SetBsPort(27012)
		server.SetWebSocketPort(27112)
		if title != common.GameAoE1 {
			server.SetName("localhost")
			server.SetOutOfBandPort(27212)
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
	var avatarStats i.A
	if title == common.GameAoE1 {
		response = append(response, i.A{})
	} else {
		avatarStats = u.EncodeAvatarStats()
	}
	allProfileInfo := i.A{
		0,
		profileInfo,
		relationship.Relationships(title, req.ClientLibVersion, users, u),
		extraProfileInfoList,
		avatarStats,
		nil,
		i.A{},
		nil,
		1,
	}
	if title != common.GameAoE1 {
		allProfileInfo = append(allProfileInfo, i.A{})
	}
	if req.ClientLibVersion >= 193 {
		allProfileInfo = append(allProfileInfo, -1)
	}
	response = append(response,
		game.Resources().LoginData(),
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
