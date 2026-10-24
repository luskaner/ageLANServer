package party

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

// reportMatchResponse is the payload the AoE2:DE client expects from
// /game/party/reportMatch. It has been verified against a real 1.16.0 match, so
// keep it byte-for-byte compatible.
func reportMatchResponse() i.A {
	return i.A{2, i.A{}, i.A{-1, "", "", "", 0, ""}, i.A{-1, "", "", "", 0, ""}, nil, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, 0, nil, i.A{}, i.A{}, i.A{}}
}

func ReportMatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logger.Printf("party.reportMatch: failed to parse form: %s\n", err)
		i.JSON(&w, reportMatchResponse())
		return
	}
	if len(r.PostForm) == 0 {
		logger.Printf("party.reportMatch: no form data\n")
		i.JSON(&w, reportMatchResponse())
		return
	}

	form := r.PostForm
	matchID, err := strconv.ParseInt(strings.TrimSpace(form.Get("match_id")), 10, 64)
	if err != nil || matchID == 0 {
		logger.Printf("party.reportMatch: invalid match_id %q\n", form.Get("match_id"))
		i.JSON(&w, reportMatchResponse())
		return
	}

	result := models.MatchResult{
		MatchID:    matchID,
		Players:    parseInt32List(form.Get("profile_ids")),
		Results:    parseInt32List(form.Get("results")),
		Teams:      parseInt32List(form.Get("teamIDs")),
		Races:      parseInt32List(form.Get("race_ids")),
		XPGained:   parseInt32List(form.Get("xpGained")),
		WinnerTeam: -1,
		FinishedAt: time.Now(),
	}
	if len(result.Players) == 0 {
		logger.Printf("party.reportMatch: match=%d has no profile_ids (%q)\n", matchID, form.Get("profile_ids"))
		i.JSON(&w, reportMatchResponse())
		return
	}

	history := models.G(r).MatchHistory()
	if history == nil {
		logger.Printf("party.reportMatch: match=%d no match history store\n", matchID)
		i.JSON(&w, reportMatchResponse())
		return
	}
	ratings, rated := history.ApplyResult(result)

	users := models.G(r).Users()
	var updates []string
	for _, profile := range result.Players {
		if rating, ok := ratings[profile]; ok && rating != nil {
			// Refresh the cached rating so the next profile encode shows the
			// new value without waiting for another database read.
			if user, found := users.GetUserById(profile); found {
				user.SetRating(rating.Rating)
			}
			updates = append(updates, strconv.FormatInt(int64(profile), 10)+"="+strconv.Itoa(int(rating.Rating)))
		}
	}
	logger.Printf(
		"party.reportMatch: stored match=%d profiles=%v results=%v teams=%v races=%v rated=%t ratings=[%s]\n",
		matchID,
		result.Players,
		result.Results,
		result.Teams,
		result.Races,
		rated,
		strings.Join(updates, " "),
	)
	i.JSON(&w, reportMatchResponse())
}

// parseInt32List accepts the JSON arrays used by the client
// ("[1,2]") and, defensively, comma separated lists.
func parseInt32List(raw string) []int32 {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil
	}
	var values []int32
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values
	}
	raw = strings.Trim(raw, "[]")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values = make([]int32, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
		if err != nil {
			return nil
		}
		values = append(values, int32(value))
	}
	return values
}
