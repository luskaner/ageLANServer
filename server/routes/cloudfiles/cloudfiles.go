package cloudfiles

import (
	"fmt"
	"net/http"
	"server/files"
	"server/models"
	"strconv"
	"strings"
	"time"
)

func Cloudfiles(w http.ResponseWriter, r *http.Request) {
	key := strings.Join(strings.Split(r.URL.Path, "/")[2:], "/")
	info, exists := models.GetCredentials(r.URL.Query().Get("sig"))

	if !exists {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	filename, file, ok := files.CloudFiles.GetByKey(key)
	if ok {
		if file.Key != info.GetKey() {
			http.Error(w, "Incorrect signature", http.StatusForbidden)
			return
		}
		data, err := files.ReadCloudFile(filename)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		lengthStr := strconv.Itoa(file.Length)
		w.Header().Set("Content-Length", lengthStr)
		w.Header().Set("Content-Type", file.Type)
		w.Header().Set("Content-MD5", file.Checksum)
		w.Header().Set("Last-Modified", file.Created)
		w.Header().Set("Accept-Range", "bytes")
		w.Header().Set("ETag", file.ETag)
		w.Header().Set("Server", "Windows-Azure-Blob/1.0 Microsoft-HTTPAPI/2.0")
		w.Header().Set("x-ms-request-id", fmt.Sprintf("%d", time.Now().Unix()))
		w.Header().Set("x-ms-version", file.Version)
		w.Header().Set("x-ms-meta-filename", filename)
		w.Header().Set("x-ms-meta-ContentLength", lengthStr)
		w.Header().Set("x-ms-creation-time", file.Created)
		w.Header().Set("x-ms-lease-status", "unlocked")
		w.Header().Set("x-ms-lease-state", "available")
		w.Header().Set("x-ms-blob-type", "BlockBlob")
		w.Header().Set("x-ms-server-encrypted", "true")
		w.Header().Set("Date", time.Now().Format(time.RFC1123))
		w.Header().Set("Content-Type", file.Type)
		_, _ = w.Write(data)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}
