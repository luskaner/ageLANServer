package Automatch2

import (
	"net/http"

	i "github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/models"
)

func GetAutomatchMap(w http.ResponseWriter, r *http.Request) {
	game := models.G(r)
	maps, ok := game.Resources().ArrayFiles()["automatchMaps.json"]
	if !ok {
		i.JSON(&w, maps)
		return
	}
	// The client renders every map of this payload. Filtering it here is what
	// makes the pool visible in game: without it the client would still offer
	// maps the server is not going to pick.
	filtered := models.FilterAutomatchMapsByPool(maps, game.Title(), models.AutomatchQueueID())
	i.JSON(&w, filtered)
}
