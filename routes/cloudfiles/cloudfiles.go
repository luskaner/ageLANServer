package cloudfiles

import (
	"aoe2DELanServer/files"
	"aoe2DELanServer/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func Cloudfiles(c *gin.Context) {
	key := c.Param("key")[1:]
	info, exists := models.GetCredentials(c.Query("sig"))

	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	signatureKey := info.GetKey()
	for filename, file := range files.CloudFiles {
		if file.Key == key {
			if file.Key != signatureKey {
				c.Status(http.StatusForbidden)
				return
			}
			lengthStr := strconv.Itoa(file.Length)
			c.Header("Content-Length", lengthStr)
			c.Header("Content-Type", file.Type)
			c.Header("Content-MD5", file.Checksum)
			c.Header("Last-Modified", file.Created)
			c.Header("Accept-Range", "bytes")
			c.Header("ETag", file.ETag)
			c.Header("Server", "Windows-Azure-Blob/1.0 Microsoft-HTTPAPI/2.0")
			c.Header("x-ms-request-id", fmt.Sprintf("%d", time.Now().Unix()))
			c.Header("x-ms-version", file.Version)
			c.Header("x-ms-meta-filename", filename)
			c.Header("x-ms-meta-ContentLength", lengthStr)
			c.Header("x-ms-creation-time", file.Created)
			c.Header("x-ms-lease-status", "unlocked")
			c.Header("x-ms-lease-state", "available")
			c.Header("x-ms-blob-type", "BlockBlob")
			c.Header("x-ms-server-encrypted", "true")
			c.Header("Date", time.Now().Format(time.RFC1123))
			c.Data(http.StatusOK, file.Type, files.Cloud[filename])
			return
		}
	}

	c.Status(http.StatusNotFound)
}
