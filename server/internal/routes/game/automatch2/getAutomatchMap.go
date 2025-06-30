package Automatch2

import (
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetAutomatchMap(w http.ResponseWriter, r *http.Request) {
	models.G(r).Resources().ReturnArrayFile("automatchMaps.json", &w)
}
