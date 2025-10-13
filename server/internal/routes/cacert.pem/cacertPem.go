package cacert_pem

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CacertPem(w http.ResponseWriter, r *http.Request) {
	exe, err := os.Executable()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	folder := common.CertificatePairFolder(exe)
	if folder == "" {
		http.NotFound(w, r)
		return
	} else {
		var file string
		if title := models.G(r).Title(); title == common.GameAoM || title == common.GameAoE4 {
			file = common.CACert
		} else {
			file = common.SelfSignedCert
		}
		path := filepath.Join(folder, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.ServeFile(w, r, path)
		}
	}
}
