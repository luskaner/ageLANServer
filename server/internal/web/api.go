package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/luskaner/ageServer/server/internal/models"
)

// writeJSON encodes a value with the status code.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

// sortedGames returns the initialized games in a stable order, with age2 first
// because it is the one this dashboard was built around.
func (s *Server) sortedGames() []string {
	ids := make([]string, 0, len(s.opts.Games))
	for id := range s.opts.Games {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(a, b int) bool {
		if ids[a] == "age2" {
			return true
		}
		if ids[b] == "age2" {
			return false
		}
		return ids[a] < ids[b]
	})
	return ids
}

// game resolves the ?game= parameter, defaulting to the first initialized game.
func (s *Server) game(r *http.Request) (models.Game, string, bool) {
	requested := strings.TrimSpace(r.URL.Query().Get("game"))
	if requested != "" {
		game, ok := s.opts.Games[requested]
		return game, requested, ok && game != nil
	}
	for _, id := range s.sortedGames() {
		if game := s.opts.Games[id]; game != nil {
			return game, id, true
		}
	}
	return nil, "", false
}

// HandleInfo is GET /api/info.
func (s *Server) HandleInfo(w http.ResponseWriter, r *http.Request) {
	type gameInfo struct {
		Game                string             `json:"game"`
		Hosts               []string           `json:"hosts"`
		Lobbies             int                `json:"lobbies"`
		OnlinePlayers       int                `json:"onlinePlayers"`
		Queue               int                `json:"queue"`
		LiveMatches         int                `json:"liveMatches"`
		BattleServers       []string           `json:"battleServers"`
		BattleServerDetails []BattleServerView `json:"battleServerDetails"`
		MapPool             *models.MapPool    `json:"mapPool,omitempty"`
	}
	games := make([]gameInfo, 0, len(s.opts.Games))
	for _, id := range s.sortedGames() {
		game := s.opts.Games[id]
		if game == nil {
			continue
		}
		info := gameInfo{
			Game:  id,
			Hosts: s.opts.GameHosts[id],
		}
		if advs := game.Advertisements(); advs != nil {
			info.Lobbies = len(advs.Snapshot(id))
		}
		info.OnlinePlayers = len(models.OnlinePlayers(game))
		if matchmaking := game.Matchmaking(); matchmaking != nil {
			snapshot := matchmaking.Snapshot()
			info.Queue = len(snapshot.Queue)
			info.LiveMatches = len(snapshot.Matches)
		}
		info.BattleServers = battleServerNames(game.BattleServers())
		info.BattleServerDetails = s.battleServerViews(id, game.BattleServers())
		if pool, ok := models.MapPools(id).Get(models.AutomatchQueueID()); ok {
			info.MapPool = pool
		}
		games = append(games, info)
	}
	uptime := int(time.Since(s.startedAt).Seconds())
	writeJSON(w, http.StatusOK, map[string]any{
		"name":          s.opts.Config.PublicTitle,
		"startedAt":     s.startedAt,
		"uptimeSeconds": uptime,
		"now":           time.Now(),
		"addresses":     s.Addresses(),
		"games":         games,
		"queue":         models.AutomatchQueueID(),
		"ladder":        models.LeaderboardForQueue(models.AutomatchQueueID()),
	})
}

// HandleBattleServers lists the battle servers currently loaded by the game
// server, including static entries and servers discovered by Battle Server
// Manager.
func (s *Server) HandleBattleServers(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	servers := s.battleServerViews(gameID, game.BattleServers())
	writeJSON(w, http.StatusOK, map[string]any{
		"game":          gameID,
		"count":         len(servers),
		"battleServers": servers,
	})
}

// HandleLobbies is GET /api/lobbies.
func (s *Server) HandleLobbies(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	lobbies := game.Advertisements().Snapshot(gameID)
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"count":   len(lobbies),
		"lobbies": lobbies,
	})
}

// HandleLobby is GET /api/lobbies/{id}.
func (s *Server) HandleLobby(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid lobby id")
		return
	}
	requestedGame := strings.TrimSpace(r.URL.Query().Get("game"))
	for _, gameID := range s.sortedGames() {
		if requestedGame != "" && gameID != requestedGame {
			continue
		}
		game := s.opts.Games[gameID]
		if game == nil {
			continue
		}
		for _, lobby := range game.Advertisements().Snapshot(gameID) {
			if lobby.ID == int32(id) {
				writeJSON(w, http.StatusOK, lobby)
				return
			}
		}
	}
	writeJSONError(w, http.StatusNotFound, "lobby not found")
}

// HandleQueue is GET /api/queue: who is waiting and the live pairings.
func (s *Server) HandleQueue(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	var snapshot models.MatchmakingSnapshot
	if matchmaking := game.Matchmaking(); matchmaking != nil {
		snapshot = matchmaking.Snapshot()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"queue":   snapshot.Queue,
		"matches": snapshot.Matches,
	})
}

// HandleMatches is GET /api/matches?limit=&page=&size=.
func (s *Server) HandleMatches(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	limit := intParam(r, "limit", 25, 1, 200)
	matches := game.MatchHistory().RecentMatches(limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"count":   len(matches),
		"matches": s.enrichMatches(game, matches),
	})
}

// HandleMatch is GET /api/matches/{id}.
func (s *Server) HandleMatch(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid match id")
		return
	}
	match, found := game.MatchHistory().FindMatch(id)
	if !found {
		writeJSONError(w, http.StatusNotFound, "match not found")
		return
	}
	enriched := s.enrichMatches(game, []models.MatchResult{match})
	writeJSON(w, http.StatusOK, map[string]any{
		"game":  gameID,
		"match": enriched[0],
	})
}

// enriches matches with the alias of every profile involved.
func (s *Server) enrichMatches(game models.Game, matches []models.MatchResult) []MatchView {
	return matchViews(game, matches)
}

// HandleLeaderboard is GET /api/leaderboard?start=&count=&game=.
func (s *Server) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	start := intParam(r, "start", 0, 0, 1_000_000)
	count := intParam(r, "count", 50, 1, 200)
	leaderboard := int32(intParam(r, "ladder", int(models.LeaderboardForQueue(models.AutomatchQueueID())), 1, 1000))
	rows, total := leaderboardRows(game, leaderboard, start, count)
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"ladder":  leaderboard,
		"start":   start,
		"count":   len(rows),
		"total":   total,
		"players": rows,
	})
}

// HandlePlayers is GET /api/players.
func (s *Server) HandlePlayers(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	players := models.OnlinePlayers(game)
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"count":   len(players),
		"players": players,
	})
}

// HandleMaps is GET /api/maps: the maps the client may choose for the queue
// plus the configured pool.
func (s *Server) HandleMaps(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	queueID := int32(intParam(r, "queue", int(models.AutomatchQueueID()), 1, 1000))
	maps := availableMaps(game, queueID)
	pool, hasPool := models.MapPools(gameID).Get(queueID)
	allowed := maps
	if hasPool {
		ids, names := models.MapPools(gameID).AllowedMaps(queueID)
		allowed = make([]models.AvailableMap, 0, len(maps))
		for _, entry := range maps {
			if _, ok := ids[entry.ID]; ok {
				allowed = append(allowed, entry)
				continue
			}
			if _, ok := names[normalizeName(entry.Name)]; ok {
				allowed = append(allowed, entry)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"game":       gameID,
		"queue":      queueID,
		"maps":       maps,
		"allowed":    allowed,
		"mapPool":    pool,
		"poolActive": hasPool,
	})
}

func availableMaps(game models.Game, queueID int32) []models.AvailableMap {
	resources := game.Resources()
	if resources == nil {
		return nil
	}
	payload, ok := resources.ArrayFiles()["automatchMaps.json"]
	if !ok {
		return nil
	}
	if queueID == models.AutomatchQueueID() {
		return models.AutomatchMapsForPlayers(payload, models.AutomatchQueueLeaderboardID(), 2)
	}
	return models.AvailableMaps(payload, queueID)
}

// HandleHealth is GET /api/health.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ok",
		"time":          time.Now(),
		"uptimeSeconds": int(time.Since(s.startedAt).Seconds()),
		"games":         len(s.opts.Games),
		"lastEvent":     s.lastEventID(),
	})
}

func intParam(r *http.Request, name string, fallback, min, max int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return fallback
	}
	return value
}
