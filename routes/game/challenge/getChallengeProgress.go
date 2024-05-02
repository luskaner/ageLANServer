package challenge

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/routes/game/challenge/shared"
	"net/http"
)

func GetChallengeProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, shared.GetChallengeProgressData()})
}
