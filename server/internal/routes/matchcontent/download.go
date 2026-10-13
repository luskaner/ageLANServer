package matchcontent

import (
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"

	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

// Download implements GET /api/MatchContent/{matchId}. The payload is streamed
// straight from disk, so serving a multi-megabyte savegame never loads it into
// the server's heap.
func Download(w http.ResponseWriter, r *http.Request) {
	files := models.G(r).CustomFiles()
	if files == nil || !files.Enabled() {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("matchId"), 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	cf, rc, ok := files.Open(int32(id))
	if !ok {
		logger.Warn("Custom file %d requested but not found", id)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer func() { _ = rc.Close() }()
	respondDownload(w, cf, rc)
}

// respondDownload is the only place that depends on the (still unknown) shape
// of a successful download response. Assumption A: the raw bytes with
// application/octet-stream metadata, mirroring routes/cloudfiles.
func respondDownload(w http.ResponseWriter, cf *models.CustomFile, rc io.ReadCloser) {
	w.Header().Set("Content-Type", cf.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(cf.Length))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", `attachment; filename="`+cf.Filename+`"`)
	w.Header().Set("Last-Modified", cf.Created.UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", `"`+cf.MD5+`"`)
	if raw, err := hex.DecodeString(cf.MD5); err == nil {
		w.Header().Set("Content-MD5", base64.StdEncoding.EncodeToString(raw))
	}
	_, _ = io.Copy(w, rc)
}
