package models

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	i "github.com/luskaner/ageLANServer/server/internal"
)

type MatchResult struct {
	MatchID     int64     `json:"matchId"`
	MapID       int32     `json:"mapId"`
	MapName     string    `json:"mapName"`
	Leaderboard int32     `json:"leaderboard"`
	Players     []int32   `json:"players"`
	Teams       []int32   `json:"teams"`
	Results     []int32   `json:"results"`
	WinnerTeam  int32     `json:"winnerTeam"`
	FinishedAt  time.Time `json:"finishedAt"`
	QueueType   string    `json:"queueType"`
}

type PlayerRating struct {
	ProfileID int32     `json:"profileId"`
	StatID    int32     `json:"statId"`
	Rating    int32     `json:"rating"`
	Games     int32     `json:"games"`
	Wins      int32     `json:"wins"`
	Losses    int32     `json:"losses"`
	LastMatch time.Time `json:"lastMatch"`
}

type MatchHistoryStore struct {
	mu      sync.RWMutex
	matches []MatchResult
	ratings map[int32]*PlayerRating // keyed by profileId
	path    string
}

func NewMatchHistoryStore(dataDir string) *MatchHistoryStore {
	s := &MatchHistoryStore{
		matches: make([]MatchResult, 0, 1024),
		ratings: make(map[int32]*PlayerRating),
		path:    filepath.Join(dataDir, "matchHistory.json"),
	}
	s.load()
	return s
}

func (s *MatchHistoryStore) GetStorePath() string { return s.path }

func (s *MatchHistoryStore) load() {
	if s.path == "" {
		return
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var data struct {
		Matches []MatchResult   `json:"matches"`
		Ratings map[int32]*PlayerRating `json:"ratings"`
	}
	if json.Unmarshal(b, &data) == nil {
		s.matches = data.Matches
		if data.Ratings != nil {
			s.ratings = data.Ratings
		}
	}
}

func (s *MatchHistoryStore) save() {
	if s.path == "" {
		return
	}
	data := struct {
		Matches []MatchResult   `json:"matches"`
		Ratings map[int32]*PlayerRating `json:"ratings"`
	}{Matches: s.matches, Ratings: s.ratings}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, b, 0644)
}

func (s *MatchHistoryStore) RecordMatch(m MatchResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.matches = append(s.matches, m)
	if len(s.matches) > 10000 {
		s.matches = s.matches[len(s.matches)-5000:]
	}
	s.save()
}

func (s *MatchHistoryStore) GetRatings(profiles []int32) map[int32]*PlayerRating {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[int32]*PlayerRating)
	for _, p := range profiles {
		if r, ok := s.ratings[p]; ok {
			cp := *r
			out[p] = &cp
		} else {
			out[p] = &PlayerRating{ProfileID: p, Rating: 1000}
		}
	}
	return out
}

func (s *MatchHistoryStore) GetRecentForProfile(profileID int32, limit int) []MatchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []MatchResult
	for i := len(s.matches) - 1; i >= 0 && len(out) < limit; i-- {
		m := s.matches[i]
		for _, p := range m.Players {
			if p == profileID {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

func (s *MatchHistoryStore) GetLeaderboard(limit int) []*PlayerRating {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*PlayerRating, 0, len(s.ratings))
	for _, r := range s.ratings {
		cp := *r
		list = append(list, &cp)
	}
	if len(list) == 0 {
		return list
	}
	// simple sort by rating desc
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Rating > list[i].Rating {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}

// Default rating for new players
func (s *MatchHistoryStore) GetDefaultRating() int32 { return 1000 }

// Ensure rating exists
func (s *MatchHistoryStore) EnsureRating(profileID int32, statID int32) *PlayerRating {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.ratings[profileID]; ok {
		return r
	}
	r := &PlayerRating{ProfileID: profileID, StatID: statID, Rating: 1000}
	s.ratings[profileID] = r
	s.save()
	return r
}

// Apply Elo update
func (s *MatchHistoryStore) UpdateRatings(m MatchResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(m.Players) == 0 {
		return
	}
	// Find winner
	winnerTeam := m.WinnerTeam
	if winnerTeam == -1 {
		// try from results
		for i, res := range m.Results {
			if res == 1 && i < len(m.Teams) {
				winnerTeam = m.Teams[i]
				break
			}
		}
	}
	// collect teams
	teamRatings := map[int32][]int32{}
	for i, p := range m.Players {
		if i >= len(m.Teams) {
			continue
		}
		t := m.Teams[i]
		if _, ok := s.ratings[p]; !ok {
			s.ratings[p] = &PlayerRating{ProfileID: p, Rating: 1000}
		}
		teamRatings[t] = append(teamRatings[t], p)
	}
	if len(teamRatings) < 2 {
		return
	}
	// simple average
	avg := func(team int32) float64 {
		sum := 0.0
		cnt := 0
		for _, p := range teamRatings[team] {
			if r, ok := s.ratings[p]; ok {
				sum += float64(r.Rating)
				cnt++
			}
		}
		if cnt == 0 { return 1000 }
		return sum / float64(cnt)
	}
	// get all teams
	teams := []int32{}
	for t := range teamRatings {
		teams = append(teams, t)
	}
	if len(teams) != 2 {
		return
	}
	t1, t2 := teams[0], teams[1]
	r1, r2 := avg(t1), avg(t2)
	K := float64(32)
	// expected
	e1 := 1.0 / (1.0 + math.Pow(10, (r2-r1)/400.0))
	e2 := 1.0 - e1
	var s1, s2 float64
	if winnerTeam == t1 {
		s1, s2 = 1.0, 0.0
	} else if winnerTeam == t2 {
		s1, s2 = 0.0, 1.0
	} else {
		s1, s2 = 0.5, 0.5
	}
	dr1 := int32(K * (s1 - e1))
	dr2 := int32(K * (s2 - e2))
	for _, p := range teamRatings[t1] {
		if r, ok := s.ratings[p]; ok {
			r.Rating += dr1
			r.Games++
			if s1 == 1.0 { r.Wins++ } else if s1 == 0.0 { r.Losses++ }
			r.LastMatch = time.Now()
		}
	}
	for _, p := range teamRatings[t2] {
		if r, ok := s.ratings[p]; ok {
			r.Rating += dr2
			r.Games++
			if s2 == 1.0 { r.Wins++ } else if s2 == 0.0 { r.Losses++ }
			r.LastMatch = time.Now()
		}
	}
	s.save()
}


