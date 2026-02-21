package playfab

import "github.com/luskaner/ageLANServer/server/internal/models"

type Game interface {
	models.Game
	PlayfabSessions() *MainSessions
}

type BaseGame struct {
	models.Game
	playfabSessions MainSessions
}

func (g *BaseGame) PlayfabSessions() *MainSessions {
	return &g.playfabSessions
}
