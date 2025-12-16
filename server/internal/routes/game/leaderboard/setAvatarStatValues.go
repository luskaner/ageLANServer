package leaderboard

import (
	"net/http"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

var fixedAvatarNames = map[string]mapset.Set[string]{
	common.GameAoM: mapset.NewSet[string]("STAT_GAUNTLET_REWARD_FAVOUR", "STAT_GAUNTLET_REWARD_XP"),
}

type setAvatarStatValuesRequest struct {
	AvatarStatIds i.Json[[]int32] `schema:"avatarStat_ids"`
	Values        i.Json[[]int64] `schema:"values"`
	// TODO: Implement when we know what it means
	UpdateTypes i.Json[[]int32] `schema:"updateTypes"`
}

func SetAvatarStatValues(w http.ResponseWriter, r *http.Request) {
	var req setAvatarStatValuesRequest
	if err := i.Bind(r, &req); err != nil || len(req.Values.Data) < 1 || len(req.AvatarStatIds.Data) != len(req.Values.Data) || len(req.UpdateTypes.Data) != len(req.Values.Data) {
		i.JSON(&w, i.A{2})
	}
	game := models.G(r)
	users := game.Users()
	sess := models.SessionOrPanic(r)
	u, ok := users.GetUserById(sess.GetUserId())
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	var encodedAvatarStats i.A
	var currentGameFixedAvatarNames mapset.Set[string]
	avatarStatDefinitions := game.LeaderboardDefinitions().AvatarStatDefinitions()
	data := u.GetAvatarStats().Data()
	fixedAvatarIds := mapset.NewThreadUnsafeSet[int32]()
	if currentGameFixedAvatarNames, ok = fixedAvatarNames[game.Title()]; ok {
		for name := range currentGameFixedAvatarNames.Iter() {
			if id, ok := avatarStatDefinitions.GetIdByName(name); ok {
				fixedAvatarIds.Add(id)
			}
		}
	}
	for j := 0; j < len(req.AvatarStatIds.Data); j++ {
		avatarStatId := req.AvatarStatIds.Data[j]
		if fixedAvatarIds.ContainsOne(avatarStatId) {
			continue
		}
		avatarStatValue := req.Values.Data[j]
		data.WithWriteLock(avatarStatId, func() {
			var avatarStat models.AvatarStat
			if avatarStat, ok = data.GetStat(avatarStatId); ok {
				avatarStat.UnsafeSetValue(req.Values.Data[j])
			} else {
				avatarStat = models.NewAvatarStat(avatarStatId, avatarStatValue)
				data.AddStat(avatarStat)
			}
			encodedAvatarStats = append(encodedAvatarStats, avatarStat.UnsafeEncode(u.GetProfileId()))
		})
	}
	if len(encodedAvatarStats) > 0 {
		// TODO: Likely others such as friends or people within the same party need to be notified too
		wss.SendOrStoreMessage(
			sess,
			"AvatarStatsUpdatedMessage",
			i.A{
				encodedAvatarStats,
			},
		)
		go func() {
			_ = u.GetAvatarStats().Save()
		}()
	}
	i.JSON(&w, i.A{0})
}
