package cloud

import (
	"aoe2DELanServer/asset"
	"aoe2DELanServer/j"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetFileURL(c *gin.Context) {
	namesStr := c.Query("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		c.JSON(http.StatusOK, j.A{2, j.A{}})
		return
	}
	descriptions := make(j.A, len(names))
	for i, name := range names {
		fileData := asset.CloudFiles[name]
		finalPart := fileData.Key
		descriptions[i] = j.A{
			name,
			fileData.Length,
			fileData.Id,
			"https://aoe-api.worldsedgelink.com/cloudfiles/" + finalPart,
			finalPart,
		}
	}
	c.JSON(http.StatusOK, j.A{0, descriptions})
}
