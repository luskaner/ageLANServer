package Automatch2

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetAutomatchMap(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, models.G(r).Resources().ArrayFiles["automatchMaps.json"])
}
