package Automatch2

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAutomatchMap(c *gin.Context) {
	automatchMaps := files.Config["automatch_maps.json"]
	response := make(i.A, len(automatchMaps))
	copy(response, automatchMaps)
	response = append(i.A{0}, i.A{response}...)
	c.JSON(http.StatusOK, response)
}
