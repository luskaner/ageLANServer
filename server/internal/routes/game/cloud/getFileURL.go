package cloud

import (
	"encoding/json"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetFileURL(w http.ResponseWriter, r *http.Request) {
	namesStr := r.URL.Query().Get("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{nil}})
		return
	}
	game := models.G(r)
	descriptions := make(i.A, len(names))
	gameTitle := game.Title()
	for j, name := range names {
		fileData, ok := game.Resources().CloudFiles.Value[name]
		if !ok {
			i.JSON(&w, i.A{2, i.A{nil}})
			return
		}
		finalPart := fileData.Key
		description := i.A{
			name,
			fileData.Length,
			fileData.Id,
			fmt.Sprintf("https://%s/cloudfiles/%s", r.Host, finalPart),
		}
		if gameTitle == common.GameAoE2 {
			description = append(
				description,
				finalPart,
			)
		}
		descriptions[j] = description
	}
	i.JSON(&w, i.A{0, descriptions})
}
