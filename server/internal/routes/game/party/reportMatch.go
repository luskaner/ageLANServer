package party

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
)

func ReportMatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logger.Printf("party.reportMatch: failed to parse form: %s\n", err)
		i.JSON(&w, i.A{2, i.A{}, i.A{-1, "", "", "", 0, ""}, i.A{-1, "", "", "", 0, ""}, nil, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, 0, nil, i.A{}, i.A{}, i.A{}})
		return
	}
	if len(r.PostForm) > 0 {
		logger.Printf("party.reportMatch: form=%s\n", r.PostForm.Encode())
	} else {
		logger.Printf("party.reportMatch: no form data\n")
	}
	i.JSON(&w, i.A{2, i.A{}, i.A{-1, "", "", "", 0, ""}, i.A{-1, "", "", "", 0, ""}, nil, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, 0, nil, i.A{}, i.A{}, i.A{}})
}
