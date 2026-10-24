package leaderboard

import (
	"net/http"
	"strconv"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

// GetLeaderBoard returns one page of a ladder.
//
// The client sends leaderboard_id, start (1-based), count and sortBy. The
// response follows the same positional layout as the other game endpoints:
//
//	[result, [leaderboardStats...], [profileInfo...], [extraProfileInfo...]]
//
// Each leaderboardStat is the 1v1 ladder entry
//
//	[profileId, rankTotal, drops, disputes, streak, losses, wins,
//	 communityEventID, highestRating, highestRankLevel, highestRank,
//	 lastMatchTime, rating, regionRankTotal, regionRanking, lastUpdated, statID]
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

	entries := i.A{}
	profiles := i.A{}
	extras := i.A{}
	now := time.Now().Unix()
	skipped := 0
	for _, rating := range page {
		user, ok := users.GetUserById(rating.ProfileID)
		if !ok {
			// The client pairs every leaderboard entry with a profileInfo, so an
			// entry without one would desynchronise the two arrays.
			skipped++
			continue
		}
		user.SetRating(rating.Rating)
		entries = append(entries, i.A{
			rating.ProfileID, // profileId
			rating.RankTotal, // rankTotal
			0,                // drops
			0,                // disputes
			rating.Streak,    // streak
			rating.Losses,    // losses
			rating.Wins,      // wins
			-1,               // communityEventID
			rating.HighestRating,
			0, // highestRankLevel
			0, // highestRank
			rating.LastMatch.Unix(),
			rating.Rating,
			rating.RankTotal, // regionRankTotal
			rating.Rank,      // regionRanking
			now,              // lastUpdated
			rating.StatID,
		})
		profiles = append(profiles, user.EncodeProfileInfo(clientLibVersion))
		extras = append(extras, user.EncodeExtraProfileInfo(clientLibVersion))
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
	i.JSON(&w, i.A{0, entries, profiles, extras})
}
