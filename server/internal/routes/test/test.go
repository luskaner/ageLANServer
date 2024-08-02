package test

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func Test(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{})
}
