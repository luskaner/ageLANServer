package web

import (
	"sort"
	"strconv"

	"github.com/luskaner/ageServer/server/internal/models"
)

type PerformanceView struct {
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Games   int    `json:"games"`
	Wins    int    `json:"wins"`
	Losses  int    `json:"losses"`
	WinRate int    `json:"winRate"`
}

type PlayerProfileView struct {
	ProfileID            int32             `json:"profileId"`
	Alias                string            `json:"alias"`
	Level                uint16            `json:"level,omitempty"`
	XP                   uint32            `json:"xp"`
	Rating               int32             `json:"rating"`
	HighestRating        int32             `json:"highestRating"`
	Games                int32             `json:"games"`
	Wins                 int32             `json:"wins"`
	Losses               int32             `json:"losses"`
	BestCivilization     string            `json:"bestCivilization,omitempty"`
	BestCivilizationIcon string            `json:"bestCivilizationIcon,omitempty"`
	BestMap              string            `json:"bestMap,omitempty"`
	BestMapIcon          string            `json:"bestMapIcon,omitempty"`
	Civilizations        []PerformanceView `json:"civilizations"`
	Maps                 []PerformanceView `json:"maps"`
	RecentMatches        []MatchView       `json:"recentMatches"`
}

func profilePerformanceView(stats []models.MatchPerformance, civilizations bool, gameID string) []PerformanceView {
	byName := make(map[string]PerformanceView, len(stats))
	for _, stat := range stats {
		name := stat.Name
		icon := ""
		if civilizations {
			raceID, err := strconv.ParseInt(name, 10, 32)
			if err != nil {
				continue
			}
			name = models.CivilizationName(gameID, int32(raceID))
			icon = models.CivilizationIconPath(gameID, int32(raceID))
		} else {
			icon = models.MapIconPath(name)
			name = models.DisplayMapName(name)
		}
		entry := byName[name]
		entry.Name = name
		entry.Icon = icon
		entry.Games += stat.Games
		entry.Wins += stat.Wins
		byName[name] = entry
	}
	result := make([]PerformanceView, 0, len(byName))
	for _, entry := range byName {
		entry.Losses = entry.Games - entry.Wins
		entry.WinRate = entry.Wins * 100 / entry.Games
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Games != result[j].Games {
			return result[i].Games > result[j].Games
		}
		if result[i].WinRate != result[j].WinRate {
			return result[i].WinRate > result[j].WinRate
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func bestPerformance(stats []PerformanceView) PerformanceView {
	if len(stats) == 0 {
		return PerformanceView{}
	}
	best := stats[0]
	for _, stat := range stats[1:] {
		if stat.Games < 3 {
			continue
		}
		if best.Games < 3 || stat.WinRate > best.WinRate ||
			(stat.WinRate == best.WinRate && stat.Games > best.Games) {
			best = stat
		}
	}
	return best
}

func buildPlayerProfile(game models.Game, gameID string, profileID int32) (PlayerProfileView, bool, error) {
	if game == nil || game.MatchHistory() == nil {
		return PlayerProfileView{}, false, nil
	}
	history := game.MatchHistory()
	profile, found := history.GetProfiles([]int32{profileID})[profileID]
	recent := history.GetRecentForProfile(profileID, 15)
	if !found && len(recent) == 0 {
		return PlayerProfileView{}, false, nil
	}

	view := PlayerProfileView{ProfileID: profileID}
	if found {
		view.Alias = profile.Alias
		view.Level = dashboardLevel(profile.Level)
		view.XP = profile.XP
	}
	if view.Alias == "" {
		view.Alias = "profile-" + itoa(int64(profileID))
	}
	if ratings := history.GetRatingsForLeaderboard(models.SoloRMRanked, []int32{profileID}); ratings[profileID] != nil {
		view.Rating = ratings[profileID].Rating
		view.HighestRating = ratings[profileID].HighestRating
		view.Games = ratings[profileID].Games
		view.Wins = ratings[profileID].Wins
		view.Losses = ratings[profileID].Losses
	}
	mapStats, civilizationStats, err := history.GetProfilePerformance(profileID)
	if err != nil {
		return PlayerProfileView{}, false, err
	}
	view.Maps = profilePerformanceView(mapStats, false, gameID)
	view.Civilizations = profilePerformanceView(civilizationStats, true, gameID)
	bestMap := bestPerformance(view.Maps)
	view.BestMap = bestMap.Name
	view.BestMapIcon = bestMap.Icon
	bestCivilization := bestPerformance(view.Civilizations)
	view.BestCivilization = bestCivilization.Name
	view.BestCivilizationIcon = bestCivilization.Icon
	view.RecentMatches = matchViews(game, recent)
	return view, true, nil
}
