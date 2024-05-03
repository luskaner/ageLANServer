package party

import (
	"net/http"
	i "server/internal"
)

func ReportMatch(w http.ResponseWriter, _ *http.Request) {
	// What else is needed to implement?
	i.JSON(&w, i.A{2, i.A{}, i.A{}, i.A{}, nil, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, 0, nil, i.A{}, i.A{}, i.A{}})
}
