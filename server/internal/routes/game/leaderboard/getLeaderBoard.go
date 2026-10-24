package leaderboard

import (
	"net/http"
	"strconv"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

// GetLeaderBoard returns one page of a ladder.
//
// The client sends leaderboard_id, start (1-based), count and sortBy. The
// response follows the positional layout of LeaderboardGetLeaderboardEvent:
//
//	[result, [statGroup...], [profileInfo...], [leaderboardStat...]]
//
// Each statGroup is [statGroupID, statGroupName, metaData, type, [profileID...]]
// and each leaderboardStat is produced by MainUser.EncodeLeaderboardStat.
func GetLeaderBoard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	logger.Printf("leaderboard.getLeaderBoard: query=%s\n", r.URL.RawQuery)

	leaderboardID := models.SoloRMRanked
	if raw := query.Get("leaderboard_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 32); err == nil && parsed > 0 {
			leaderboardID = int32(parsed)
		}
	}
	start := 1
	if raw := query.Get("start"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			start = parsed
		}
	}
	count := 100
	if raw := query.Get("count"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			count = parsed
		}
	}

	game := models.G(r)
	page, total := game.MatchHistory().GetLeaderboardPage(leaderboardID, start-1, count)
	users := game.Users()
	clientLibVersion := models.SessionOrPanic(r).GetClientLibVersion()

	statGroups := i.A{}
	profiles := i.A{}
	stats := i.A{}
	skipped := 0
	for _, rating := range page {
		user, ok := users.GetUserById(rating.ProfileID)
		if !ok {
			// Every stat group must have its profile info, otherwise the client
			// cannot resolve the entry.
			skipped++
			continue
		}
		user.SetRating(rating.Rating)
		statGroups = append(statGroups, i.A{
			user.GetStatGroupID(), // statGroupID
			"",               // statGroupName
			"",               // metaData
			1,                // type
			i.A{user.GetId()},
		})
		profiles = append(profiles, user.EncodeProfileInfo(clientLibVersion))
		stats = append(stats, user.EncodeLeaderboardStat(rating))
	}
	logger.Printf(
		"leaderboard.getLeaderBoard: leaderboard=%d start=%d count=%d total=%d returned=%d skipped=%d profiles=%d\n",
		leaderboardID,
		start,
		count,
		total,
		len(page),
		skipped,
		len(profiles),
	)
	i.JSON(&w, i.A{0, statGroups, profiles, stats})
}
