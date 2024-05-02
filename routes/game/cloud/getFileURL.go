package cloud

import (
	"aoe2DELanServer/files"
	i "aoe2DELanServer/internal"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetFileURL(c *gin.Context) {
	namesStr := c.Query("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		c.JSON(http.StatusOK, i.A{2, i.A{}})
		return
	}
	descriptions := make(i.A, len(names))
	for j, name := range names {
		fileData := files.CloudFiles[name]
		finalPart := fileData.Key
		descriptions[j] = i.A{
			name,
			fileData.Length,
			fileData.Id,
			"https://aoe-api.worldsedgelink.com/cloudfiles/" + finalPart,
			finalPart,
		}
	}
	c.JSON(http.StatusOK, i.A{0, descriptions})
}
