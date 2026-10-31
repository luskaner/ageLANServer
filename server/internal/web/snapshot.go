package web

import (
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/luskaner/ageServer/server/internal/models"
)

// LeaderboardRow is one row of the public ladder: the rating plus the identity
// the profile had when it was last seen.
type LeaderboardRow struct {
	ProfileID      int32     `json:"profileId"`
	Alias          string    `json:"alias"`
	Rank           int64     `json:"rank"`
	Rating         int32     `json:"rating"`
	HighestRating  int32     `json:"highestRating"`
	Games          int32     `json:"games"`
	Wins           int32     `json:"wins"`
	Losses         int32     `json:"losses"`
	Streak         int32     `json:"streak"`
	Level          uint16    `json:"level,omitempty"`
	PlatformUserID uint64    `json:"platformUserId"`
	LastMatch      time.Time `json:"lastMatch"`
}

// MatchPlayerView is one player of a finished match with its alias.
type MatchPlayerView struct {
	models.MatchPlayer
	Alias            string `json:"alias"`
	Civilization     string `json:"civilization"`
	CivilizationIcon string `json:"civilizationIcon"`
	Won              bool   `json:"won"`
}

// MatchView is a finished match as the dashboard shows it.
type MatchView struct {
	models.MatchResult
	DisplayMapName  string            `json:"displayMapName"`
	MapIcon         string            `json:"mapIcon"`
	PlayerInfo      []MatchPlayerView `json:"playerInfo"`
	DurationSeconds int               `json:"durationSeconds"`
}

func dashboardLevel(level uint16) uint16 {
	if level == 9_999 {
		return 0
	}
	return level
}

// BattleServerView describes a battle server currently loaded for one game.
type BattleServerView struct {
	Game          string    `json:"game"`
	Region        string    `json:"region"`
	Name          string    `json:"name"`
	IPv4          string    `json:"ipv4"`
	BsPort        int       `json:"bsPort"`
	WebSocketPort int       `json:"webSocketPort"`
	OutOfBandPort int       `json:"outOfBandPort,omitempty"`
	Source        string    `json:"source"`
	Status        string    `json:"status"`
	Active        bool      `json:"active"`
	CheckedAt     time.Time `json:"checkedAt,omitempty"`
}

const battleServerHealthCacheDuration = 8 * time.Second
const battleServerHealthProbeTimeout = 350 * time.Millisecond

type battleServerHealth struct {
	status    string
	active    bool
	checkedAt time.Time
	probing   bool
}

// matchViews joins match results with the stored profile aliases.
func matchViews(game models.Game, matches []models.MatchResult) []MatchView {
	profiles := make(map[int32]*models.StoredProfile)
	ids := make([]int32, 0, len(matches)*2)
	for _, match := range matches {
		ids = append(ids, match.Players...)
	}
	if history := game.MatchHistory(); history != nil && len(ids) > 0 {
		profiles = history.GetProfiles(ids)
	}
	views := make([]MatchView, 0, len(matches))
	for _, match := range matches {
		view := MatchView{
			MatchResult:    match,
			DisplayMapName: models.DisplayMapName(match.MapName),
			MapIcon:        models.MapIconPath(match.MapName),
		}
		if !match.FinishedAt.IsZero() && !match.StartedAt.IsZero() {
			view.DurationSeconds = int(match.FinishedAt.Sub(match.StartedAt).Seconds())
		}
		for _, player := range match.PlayerStats {
			entry := MatchPlayerView{
				MatchPlayer:      player,
				Civilization:     models.CivilizationName(game.Title(), player.RaceID),
				CivilizationIcon: models.CivilizationIconPath(game.Title(), player.RaceID),
				Won:              player.Result == 1,
			}
			if profile, ok := profiles[player.ProfileID]; ok {
				entry.Alias = profile.Alias
			}
			if entry.Alias == "" {
				entry.Alias = "profile-" + itoa(int64(player.ProfileID))
			}
			view.PlayerInfo = append(view.PlayerInfo, entry)
		}
		if len(view.PlayerInfo) == 0 {
			for index, profileID := range match.Players {
				raceID := int32(-1)
				if index < len(match.Races) {
					raceID = match.Races[index]
				}
				entry := MatchPlayerView{
					MatchPlayer:      models.MatchPlayer{ProfileID: profileID, RaceID: raceID},
					Civilization:     models.CivilizationName(game.Title(), raceID),
					CivilizationIcon: models.CivilizationIconPath(game.Title(), raceID),
					Won:              index < len(match.Results) && match.Results[index] == 1,
				}
				if profile, ok := profiles[profileID]; ok {
					entry.Alias = profile.Alias
				}
				if entry.Alias == "" {
					entry.Alias = "profile-" + itoa(int64(profileID))
				}
				view.PlayerInfo = append(view.PlayerInfo, entry)
			}
		}
		views = append(views, view)
	}
	return views
}

// leaderboardRows joins the ratings of a ladder with the stored profile
// identity. It is the single source of the ladder for the API, SSE and the
// server-rendered pages.
func leaderboardRows(game models.Game, ladder int32, start, count int) ([]LeaderboardRow, int64) {
	history := game.MatchHistory()
	if history == nil {
		return nil, 0
	}
	ratings, total := history.GetLeaderboardPage(ladder, start, count)
	ids := make([]int32, 0, len(ratings))
	for _, rating := range ratings {
		ids = append(ids, rating.ProfileID)
	}
	profiles := history.GetProfiles(ids)
	rows := make([]LeaderboardRow, 0, len(ratings))
	for _, rating := range ratings {
		row := LeaderboardRow{
			ProfileID:     rating.ProfileID,
			Rank:          rating.Rank,
			Rating:        rating.Rating,
			HighestRating: rating.HighestRating,
			Games:         rating.Games,
			Wins:          rating.Wins,
			Losses:        rating.Losses,
			Streak:        rating.Streak,
			LastMatch:     rating.LastMatch,
		}
		if profile, ok := profiles[rating.ProfileID]; ok {
			row.Alias = profile.Alias
			row.Level = dashboardLevel(profile.Level)
			row.PlatformUserID = profile.PlatformUserID
		}
		if row.Alias == "" {
			row.Alias = "profile-" + itoa(int64(rating.ProfileID))
		}
		rows = append(rows, row)
	}
	return rows, total
}

func itoa(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 20)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// GameSnapshot is the live state of one game, sent over SSE and used to render
// the pages.
type GameSnapshot struct {
	Game                string                `json:"game"`
	Lobbies             []models.PublicLobby  `json:"lobbies"`
	Players             []models.PublicPlayer `json:"players"`
	Queue               []models.QueuePlayer  `json:"queue"`
	LiveMatches         []models.LiveMatch    `json:"liveMatches"`
	Ladder              []LeaderboardRow      `json:"ladder"`
	LadderTotal         int64                 `json:"ladderTotal"`
	MapPool             *models.MapPool       `json:"mapPool,omitempty"`
	AllowedMaps         []models.AvailableMap `json:"allowedMaps,omitempty"`
	BattleServers       []string              `json:"battleServers"`
	BattleServerDetails []BattleServerView    `json:"battleServerDetails"`
	LadderID            int32                 `json:"ladderId"`
}

// Snapshot is the whole dashboard state.
type Snapshot struct {
	Time      time.Time      `json:"time"`
	EventID   uint64         `json:"eventId"`
	Games     []GameSnapshot `json:"games"`
	Uptime    int            `json:"uptimeSeconds"`
	Addresses []string       `json:"addresses"`
}

func (s *Server) battleServerViews(gameID string, servers models.BattleServers) []BattleServerView {
	if servers == nil {
		return nil
	}
	var views []BattleServerView
	for region, server := range servers.Iter() {
		if server == nil {
			continue
		}
		entry := BattleServerView{
			Game:   gameID,
			Region: region,
			Source: "Dynamic (config, Manager or embedded)",
			Status: "Checking",
		}
		if _, isStatic := s.staticBattleServers[gameID][region]; isStatic {
			entry.Source = "Static"
		}
		if main, ok := server.(*models.MainBattleServer); ok {
			entry.Name = main.Name
			entry.IPv4 = main.IPv4
			entry.BsPort = main.BsPort
			entry.WebSocketPort = main.WebSocketPort
			entry.OutOfBandPort = main.OutOfBandPort
			entry.Active, entry.Status, entry.CheckedAt = s.battleServerHealth(
				gameID, region, main.IPv4, main.BsPort, main.WebSocketPort,
			)
		} else {
			entry.Name = server.String()
			entry.Status = "Status unavailable"
		}
		views = append(views, entry)
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Region < views[j].Region })
	return views
}

func battleServerNames(servers models.BattleServers) []string {
	if servers == nil {
		return nil
	}
	var names []string
	for region, server := range servers.Iter() {
		if server == nil {
			continue
		}
		name := region
		if details := server.String(); details != "" {
			name += ": " + details
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Server) battleServerHealth(gameID, region, host string, bsPort, webSocketPort int) (bool, string, time.Time) {
	key := fmt.Sprintf("%s/%s/%s/%d/%d", gameID, region, host, bsPort, webSocketPort)
	now := time.Now()
	s.battleHealthMu.Lock()
	cached := s.battleHealth[key]
	if cached.checkedAt.IsZero() || now.Sub(cached.checkedAt) >= battleServerHealthCacheDuration {
		if !cached.probing {
			cached.probing = true
			s.battleHealth[key] = cached
			go s.probeBattleServer(key, host, bsPort, webSocketPort)
		}
	}
	s.battleHealthMu.Unlock()
	if cached.checkedAt.IsZero() {
		return false, "Checking", time.Time{}
	}
	if cached.probing && now.Sub(cached.checkedAt) >= battleServerHealthCacheDuration {
		return cached.active, "Checking", cached.checkedAt
	}
	return cached.active, cached.status, cached.checkedAt
}

func (s *Server) probeBattleServer(key, host string, bsPort, webSocketPort int) {
	if host == "" || host == "auto" {
		host = "127.0.0.1"
	}
	type result struct {
		port int
		ok   bool
	}
	results := make(chan result, 2)
	for _, port := range []int{bsPort, webSocketPort} {
		go func(port int) {
			if port < 1 || port > 65535 {
				results <- result{port: port}
				return
			}
			conn, err := net.DialTimeout("tcp4", net.JoinHostPort(host, fmt.Sprint(port)), battleServerHealthProbeTimeout)
			if err == nil {
				_ = conn.Close()
			}
			results <- result{port: port, ok: err == nil}
		}(port)
	}
	bsOpen, webSocketOpen := false, false
	for range 2 {
		result := <-results
		if result.port == bsPort {
			bsOpen = result.ok
		}
		if result.port == webSocketPort {
			webSocketOpen = result.ok
		}
	}
	status := "Offline"
	if bsOpen && webSocketOpen {
		status = "Online"
	} else if bsOpen || webSocketOpen {
		status = "Partially reachable"
	}
	s.battleHealthMu.Lock()
	s.battleHealth[key] = battleServerHealth{
		status: status, active: bsOpen && webSocketOpen, checkedAt: time.Now(),
	}
	s.battleHealthMu.Unlock()
}

// BuildSnapshot collects the current state of every initialized game.
func (s *Server) BuildSnapshot() Snapshot {
	snapshot := Snapshot{
		Time:      time.Now(),
		EventID:   s.lastEventID(),
		Uptime:    int(time.Since(s.startedAt).Seconds()),
		Addresses: s.Addresses(),
	}
	queueID := models.AutomatchQueueID()
	ladder := models.LeaderboardForQueue(queueID)
	for _, id := range s.sortedGames() {
		game := s.opts.Games[id]
		if game == nil {
			continue
		}
		state := GameSnapshot{Game: id, LadderID: ladder}
		if advs := game.Advertisements(); advs != nil {
			state.Lobbies = advs.Snapshot(id)
		}
		state.Players = models.OnlinePlayers(game)
		state.BattleServers = battleServerNames(game.BattleServers())
		state.BattleServerDetails = s.battleServerViews(id, game.BattleServers())
		if matchmaking := game.Matchmaking(); matchmaking != nil {
			matchmakingSnapshot := matchmaking.Snapshot()
			state.Queue = matchmakingSnapshot.Queue
			state.LiveMatches = matchmakingSnapshot.Matches
		}
		state.Ladder, state.LadderTotal = leaderboardRows(game, ladder, 0, 10)
		if pool, ok := models.MapPools(id).Get(queueID); ok {
			state.MapPool = pool
		}
		state.AllowedMaps = availableMaps(game, queueID)
		snapshot.Games = append(snapshot.Games, state)
	}
	return snapshot
}
