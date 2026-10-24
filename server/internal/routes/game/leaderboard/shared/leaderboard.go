package shared

import (
	"encoding/json"
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func GetStatGroups(r *http.Request, idsQuery string, isProfileId bool, includeLeaderboardStat bool) i.A {
	var ids []int32
	err := json.Unmarshal([]byte(idsQuery), &ids)
	if err != nil {
		return nil
	}
	message := i.A{
		0,     // resultCode: 0 = success, 2 = error
		i.A{}, // statGroups
		i.A{}, // profileInfos
		i.A{}, // leaderboardStats
	}
	game := models.G(r)
	users := game.Users()
	clientLibVersion := models.SessionOrPanic(r).GetClientLibVersion()
	leaderboardID := models.SoloRMRanked
	if raw := r.URL.Query().Get("leaderboard_id"); raw != "" {
		if parsed, parseErr := strconv.ParseInt(raw, 10, 32); parseErr == nil && parsed > 0 {
			leaderboardID = int32(parsed)
		}
	}
	ratings := game.MatchHistory().GetRatingsForLeaderboard(leaderboardID, ids)
	for _, id := range ids {
		var u models.User
		var ok bool
		if isProfileId {
			u, ok = users.GetUserById(id)
		} else {
			u, ok = users.GetUserByStatGroupID(id)
		}
		if !ok {
			continue
		}
		if rating, found := ratings[u.GetId()]; found {
			u.SetRating(rating.Rating)
		}
		message[1] = append(message[1].(i.A), i.A{
			u.GetStatGroupID(),
			"",             // statGroupName: empty = no name
			"",             // metaData: empty = none
			1,              // type
			i.A{u.GetId()}, // profileIDs
		})
		message[2] = append(message[2].(i.A), u.EncodeProfileInfo(clientLibVersion))
		if includeLeaderboardStat {
			message[3] = append(message[3].(i.A), u.EncodeLeaderboardStat(clientLibVersion))
		}
		break
	}

	return message
}
