package cloud

import (
	"encoding/json"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/files"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetFileURL(w http.ResponseWriter, r *http.Request) {
	namesStr := r.URL.Query().Get("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
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
			"https://" + common.Domain + "/cloudfiles/" + finalPart,
			finalPart,
		}
	}
	i.JSON(&w, i.A{0, descriptions})
}
