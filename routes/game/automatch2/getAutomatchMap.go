package Automatch2

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/j"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAutomatchMap(c *gin.Context) {
	automatchMaps := asset.Files["automatch_maps.json"]
	response := make(j.A, len(automatchMaps))
	copy(response, automatchMaps)
	response = append(j.A{0}, j.A{response}...)
	c.JSON(http.StatusOK, response)
}
