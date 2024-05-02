package Automatch2

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetAutomatchMap(w http.ResponseWriter, _ *http.Request) {
	automatchMaps := files.Config["automatch_maps.json"]
	response := make(i.A, len(automatchMaps))
	copy(response, automatchMaps)
	response = append(i.A{0}, i.A{response}...)
	i.JSON(&w, response)
}
