package news

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetNews(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
