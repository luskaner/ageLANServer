package challenge

import (
	"net/http"

	i "github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/routes/game/challenge/shared"
)

func GetChallengeProgress(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, shared.GetChallengeProgressData()})
}
