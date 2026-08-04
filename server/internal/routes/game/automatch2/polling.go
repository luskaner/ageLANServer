package Automatch2

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func Polling(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement for matchmaking
	i.JSON(&w, i.A{
		2,
		i.A{},
		-1,
		nil,
		-1,
		"",
		-1,
		-1,
		-1,
		"0",
		"0",
		"-1",
		-1,
		"authtoken",
		"0.0.0.0",
		0,
		0,
		0,
		"",
		i.A{},
		nil,
		0,
	})
}
