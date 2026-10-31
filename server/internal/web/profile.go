package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/luskaner/ageServer/server/internal/models"
)

func (s *Server) playerProfile(game models.Game, gameID string, profileID int32) (PlayerProfileView, bool, error) {
	return buildPlayerProfile(game, gameID, profileID)
}

func parseProfileID(raw string) (int32, error) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid profile id")
	}
	return int32(id), nil
}

func (s *Server) HandlePlayerProfile(w http.ResponseWriter, r *http.Request) {
	game, gameID, ok := s.game(r)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown game")
		return
	}
	profileID, err := parseProfileID(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	profile, found, err := s.playerProfile(game, gameID, profileID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "player profile not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"game":    gameID,
		"profile": profile,
	})
}

func (s *Server) handlePlayerProfilePage(w http.ResponseWriter, r *http.Request) {
	profileID, err := parseProfileID(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	game, gameID, ok := s.game(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	profile, found, err := s.playerProfile(game, gameID, profileID)
	if err != nil {
		http.Error(w, "could not load player profile", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	page := s.basePage(r, profile.Alias, "/players/"+r.PathValue("id"))
	page.GameID = gameID
	page.PlayerProfile = &profile
	s.executePage(w, "profile", page)
}
