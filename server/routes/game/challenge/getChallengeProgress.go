package challenge

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/routes/game/challenge/shared"
	"net/http"
)

func GetChallengeProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, shared.GetChallengeProgressData()})
}
