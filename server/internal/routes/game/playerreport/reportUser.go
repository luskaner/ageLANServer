package playerreport

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func ReportUser(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, 0})
}
