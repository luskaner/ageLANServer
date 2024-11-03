package challenge

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/challenge/shared"
	"net/http"
)

func GetChallengeProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, shared.GetChallengeProgressData()})
}
