package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/luskaner/ageServer/server/internal/events"
	"github.com/luskaner/ageServer/server/internal/models"
)

// HandleAdminState is GET /admin/api/state: everything the panel shows in one
// request.
func (s *Server) HandleAdminState(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	type gameState struct {
		Game          string                `json:"game"`
		Hosts         []string              `json:"hosts"`
		Lobbies       []models.PublicLobby  `json:"lobbies"`
		Players       []models.PublicPlayer `json:"players"`
		Queue         []models.QueuePlayer  `json:"queue"`
		Matches       []models.LiveMatch    `json:"matches"`
		Recent        []MatchView           `json:"recentMatches"`
		BattleServers []BattleServerView    `json:"battleServers"`
		Pool          *models.MapPool       `json:"mapPool"`
		Available     []models.AvailableMap `json:"availableMaps"`
	}
	games := make([]gameState, 0, len(s.opts.Games))
	for _, id := range s.sortedGames() {
		game := s.opts.Games[id]
		if game == nil {
			continue
		}
		state := gameState{Game: id, Hosts: s.opts.GameHosts[id]}
		state.Lobbies = game.Advertisements().Snapshot(id)
		state.Players = models.OnlinePlayers(game)
		state.BattleServers = s.battleServerViews(id, game.BattleServers())
		if matchmaking := game.Matchmaking(); matchmaking != nil {
			snapshot := matchmaking.Snapshot()
			state.Queue = snapshot.Queue
			state.Matches = snapshot.Matches
		}
		state.Recent = s.enrichMatches(game, game.MatchHistory().RecentMatches(10))
		queueID := models.AutomatchQueueID()
		if pool, ok := models.MapPools(id).Get(queueID); ok {
			state.Pool = pool
		}
		state.Available = availableMaps(game, queueID)
		games = append(games, state)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"games": games,
		"csrf":  session.csrf,
	})
}

// HandleAdminPool is GET /admin/api/pool (read) and PUT /admin/api/pool
// (replace).
func (s *Server) HandleAdminPool(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	if r.Method == http.MethodGet {
		game, gameID, ok := s.game(r)
		if !ok {
			writeJSONError(w, http.StatusNotFound, "unknown game")
			return
		}
		queueID := int32(intParam(r, "queue", int(models.AutomatchQueueID()), 1, 1000))
		pool, hasPool := models.MapPools(gameID).Get(queueID)
		writeJSON(w, http.StatusOK, map[string]any{
			"game":       gameID,
			"queue":      queueID,
			"mapPool":    pool,
			"poolActive": hasPool,
			"available":  availableMaps(game, queueID),
		})
		return
	}

	var request struct {
		Game   string                `json:"game"`
		Queue  int32                 `json:"queue"`
		Mode   string                `json:"mode"`
		Cursor int                   `json:"cursor"`
		Maps   []models.MapPoolEntry `json:"maps"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Queue <= 0 {
		request.Queue = models.AutomatchQueueID()
	}
	game, ok := s.opts.Games[request.Game]
	if !ok || game == nil {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	if request.Queue > 1000 {
		writeJSONError(w, http.StatusBadRequest, "queue must be between 1 and 1000")
		return
	}
	available := availableMaps(game, request.Queue)
	availableByID := make(map[int32]models.AvailableMap, len(available))
	for _, entry := range available {
		availableByID[entry.ID] = entry
	}
	for index := range request.Maps {
		entry := &request.Maps[index]
		known, exists := availableByID[entry.MapID]
		if !exists {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("map id %d is not available for this queue", entry.MapID))
			return
		}
		entry.MapID = known.ID
		entry.Name = known.Name
	}
	pool := models.MapPool{
		GameID:  request.Game,
		QueueID: request.Queue,
		Mode:    request.Mode,
		Cursor:  request.Cursor,
		Maps:    request.Maps,
	}
	if err := models.MapPools(request.Game).Set(pool); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	stored, _ := models.MapPools(request.Game).Get(request.Queue)
	writeJSON(w, http.StatusOK, map[string]any{"mapPool": stored})
}

// HandleAdminPoolReset is POST /admin/api/pool/reset: remove the pool, so the
// matchmaker goes back to every non-vetoed map.
func (s *Server) HandleAdminPoolReset(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var request struct {
		Game  string `json:"game"`
		Queue int32  `json:"queue"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Queue <= 0 {
		request.Queue = models.AutomatchQueueID()
	}
	if request.Queue > 1000 {
		writeJSONError(w, http.StatusBadRequest, "queue must be between 1 and 1000")
		return
	}
	if game := s.opts.Games[request.Game]; game == nil {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	if err := models.MapPools(request.Game).Reset(request.Queue); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleAdminPoolCursor is POST /admin/api/pool/cursor: move the rotation
// cursor to a given map.
func (s *Server) HandleAdminPoolCursor(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var request struct {
		Game   string `json:"game"`
		Queue  int32  `json:"queue"`
		Cursor int    `json:"cursor"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Queue <= 0 {
		request.Queue = models.AutomatchQueueID()
	}
	if request.Queue > 1000 {
		writeJSONError(w, http.StatusBadRequest, "queue must be between 1 and 1000")
		return
	}
	if game := s.opts.Games[request.Game]; game == nil {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	if err := models.MapPools(request.Game).SetCursor(request.Queue, request.Cursor); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	stored, _ := models.MapPools(request.Game).Get(request.Queue)
	writeJSON(w, http.StatusOK, map[string]any{"mapPool": stored})
}

// HandleAdminPassword is POST /admin/api/password.
func (s *Server) HandleAdminPassword(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	var request struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	accepted, err := s.passwordAccepted(request.Current)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !accepted {
		writeJSONError(w, http.StatusForbidden, "the current password is not valid")
		return
	}
	if len(request.New) < 8 {
		writeJSONError(w, http.StatusBadRequest, "the new password needs at least 8 characters")
		return
	}
	hash, err := hashPassword(request.New)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err = s.store.Set("admin_password_hash", hash); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Every other browser is logged out; this one keeps its session.
	s.sessions.mu.Lock()
	for token, other := range s.sessions.sessions {
		if other != session {
			delete(s.sessions.sessions, token)
		}
	}
	s.sessions.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleAdminConfig is GET and POST /admin/api/config.
func (s *Server) HandleAdminConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	fields := ConfigFields()
	if r.Method == http.MethodGet {
		content, err := os.ReadFile(s.opts.ConfigFile)
		exists := err == nil
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		values := map[string]any{}
		if exists {
			parsed, parseErr := ParseConfig(content)
			if parseErr != nil {
				writeJSONError(w, http.StatusInternalServerError, parseErr.Error())
				return
			}
			values = parsed
		}
		values["Web.PublicTitle"] = s.opts.Config.PublicTitle
		writeJSON(w, http.StatusOK, map[string]any{
			"path":            s.opts.ConfigFile,
			"exists":          exists,
			"editable":        s.opts.ConfigFile != "",
			"fields":          fields,
			"values":          values,
			"raw":             string(content),
			"restartRequired": true,
		})
		return
	}

	var request struct {
		Changes map[string]any `json:"changes"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(request.Changes) == 0 {
		writeJSONError(w, http.StatusBadRequest, "no changes")
		return
	}
	content, err := os.ReadFile(s.opts.ConfigFile)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	literals := make(map[string]string, len(request.Changes))
	var publicTitle *string
	for path, value := range request.Changes {
		field, ok := findField(fields, path)
		if !ok {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("%q is not editable", path))
			return
		}
		if path == "Web.PublicTitle" {
			title, ok := value.(string)
			if !ok {
				writeJSONError(w, http.StatusBadRequest, "site title must be a string")
				return
			}
			title = strings.TrimSpace(title)
			if title == "" || len([]rune(title)) > 64 {
				writeJSONError(w, http.StatusBadRequest, "site title must contain 1 to 64 characters")
				return
			}
			value = title
			publicTitle = &title
		}
		literal, formatErr := FormatTomlValue(field, value)
		if formatErr != nil {
			writeJSONError(w, http.StatusBadRequest, formatErr.Error())
			return
		}
		literals[path] = literal
	}
	updated, err := ApplyConfigEdits(content, fields, literals)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err = SaveConfigFile(s.opts.ConfigFile, updated); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if publicTitle != nil {
		if err = s.store.Set("public_title", *publicTitle); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.opts.Config.PublicTitle = *publicTitle
	}
	events.Publish("config.updated", "", map[string]any{"path": s.opts.ConfigFile})
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"path":            s.opts.ConfigFile,
		"restartRequired": true,
	})
}

// HandleAdminConfigRaw is POST /admin/api/config/raw: replace the whole file,
// after validating that it parses.
func (s *Server) HandleAdminConfigRaw(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	var request struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.opts.ConfigFile == "" {
		writeJSONError(w, http.StatusConflict, "no config file is in use")
		return
	}
	parsed, err := ParseConfig([]byte(request.Content))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid TOML: "+err.Error())
		return
	}
	var publicTitle *string
	if titleValue, exists := parsed["Web.PublicTitle"]; exists {
		title, ok := titleValue.(string)
		if !ok {
			writeJSONError(w, http.StatusBadRequest, "site title must be a string")
			return
		}
		title = strings.TrimSpace(title)
		if title == "" || len([]rune(title)) > 64 {
			writeJSONError(w, http.StatusBadRequest, "site title must contain 1 to 64 characters")
			return
		}
		publicTitle = &title
	}
	if _, err := SaveConfigFile(s.opts.ConfigFile, []byte(request.Content)); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if publicTitle != nil {
		if err = s.store.Set("public_title", *publicTitle); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.opts.Config.PublicTitle = *publicTitle
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restartRequired": true})
}

// HandleAdminCloseLobby is POST /admin/api/lobbies/{id}/close.
func (s *Server) HandleAdminCloseLobby(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid lobby id")
		return
	}
	advertisements := game.Advertisements()
	var peers []int32
	foundLobby := false
	advertisements.WithWriteLock(int32(id), func() {
		adv, found := advertisements.GetAdvertisement(int32(id))
		if !found {
			return
		}
		foundLobby = true
		if main, ok := adv.(*models.MainAdvertisement); ok {
			_, peersIter := main.GetPeers().Values()
			for peer := range peersIter {
				peers = append(peers, peer.GetUserId())
			}
		}
		advertisements.UnsafeDelete(adv)
	})
	if !foundLobby {
		writeJSONError(w, http.StatusNotFound, "lobby not found")
		return
	}
	if matchmaking := game.Matchmaking(); matchmaking != nil {
		for _, peer := range peers {
			matchmaking.RemoveUser(peer)
		}
	}
	events.Publish("lobby.closed", gameID, map[string]any{"id": id, "by": "admin"})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// HandleAdminKick is POST /admin/api/queue/{user}/kick.
func (s *Server) HandleAdminKick(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	userID, err := strconv.ParseInt(r.PathValue("user"), 10, 32)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if matchmaking := game.Matchmaking(); matchmaking != nil {
		matchmaking.RemoveUser(int32(userID))
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "game": gameID})
}

// HandleAdminMatchesReload is a no-op kept for symmetry with the panel's
// refresh button; it returns the newest matches.
func (s *Server) HandleAdminRecentMatches(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"matches": s.enrichMatches(game, game.MatchHistory().RecentMatches(intParam(r, "limit", 20, 1, 100))),
	})
}

// HandleAdminBattleServers lists and saves persistent, fixed-address battle
// servers. They are loaded by the game server on its next startup.
func (s *Server) HandleAdminBattleServers(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	if r.Method == http.MethodGet {
		gameID := strings.TrimSpace(r.URL.Query().Get("game"))
		if _, ok := s.opts.Games[gameID]; !ok {
			writeJSONError(w, http.StatusNotFound, "unknown game")
			return
		}
		servers, err := models.StaticBattleServers(gameID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"game": gameID, "servers": servers})
		return
	}

	var server models.StaticBattleServer
	if err := decodeJSON(r, &server); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	gameInstance, exists := s.opts.Games[server.GameID]
	if !exists || gameInstance == nil {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	stored, err := models.StaticBattleServers(server.GameID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	isUpdate := false
	for _, current := range stored {
		if current.Region == strings.TrimSpace(server.Region) {
			isUpdate = true
			break
		}
	}
	if !isUpdate && gameInstance.BattleServers() != nil {
		if _, exists := gameInstance.BattleServers().Get(strings.TrimSpace(server.Region)); exists {
			writeJSONError(w, http.StatusConflict, "a configured or manager-discovered server already uses this region")
			return
		}
	}
	if err = models.SaveStaticBattleServer(server); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	events.Publish("battle-server.saved", server.GameID, map[string]any{"region": strings.TrimSpace(server.Region)})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restartRequired": true})
}

// HandleAdminDeleteBattleServer removes one persisted fixed-address server.
func (s *Server) HandleAdminDeleteBattleServer(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	gameID := r.PathValue("game")
	if _, ok := s.opts.Games[gameID]; !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	region := r.PathValue("region")
	if err := models.DeleteStaticBattleServer(gameID, region); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err.Error())
		return
	}
	events.Publish("battle-server.deleted", gameID, map[string]any{"region": region})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restartRequired": true})
}

func findField(fields []ConfigField, path string) (ConfigField, bool) {
	for _, field := range fields {
		if field.Path == path {
			return field, true
		}
	}
	return ConfigField{}, false
}

func normalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, ".rms2")
	name = strings.TrimSuffix(name, ".rms")
	name = strings.ReplaceAll(name, "_", " ")
	return strings.Join(strings.Fields(name), " ")
}

func decodeJSON(r *http.Request, destination any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty body")
	}
	return json.Unmarshal(body, destination)
}

// sortedKeys is a small helper used by tests and the panel.
func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
