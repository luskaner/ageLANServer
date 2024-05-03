package Automatch2

import (
	"net/http"
	"server/files"
	i "server/internal"
)

func GetAutomatchMap(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, files.ArrayFiles["automatchMaps.json"])
}
