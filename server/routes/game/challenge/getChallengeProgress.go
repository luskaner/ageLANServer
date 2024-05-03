package challenge

import (
	"net/http"
	i "server/internal"
	"server/routes/game/challenge/shared"
)

func GetChallengeProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, shared.GetChallengeProgressData()})
}
