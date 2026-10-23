package Automatch2

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/wss"
)

func Polling(w http.ResponseWriter, r *http.Request) {
	g := models.G(r)
	var req models.AutomatchPollingRequest
	if err := i.Bind(r, &req); err != nil {
		logger.Printf("automatch2: could not bind polling request: %v\n", err)
		i.JSON(&w, models.AutomatchUnmatchedResponse(0, nil))
		return
	}
	sess := models.SessionOrPanic(r)
	response, notifications := g.Matchmaking().Poll(sess, &req, r)
	i.JSON(&w, response)
	wss.SendMatchNotifications(notifications)
}
