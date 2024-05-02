package files

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/static"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

import "github.com/iancoleman/orderedmap"

var keyedFilenames = map[string]struct{}{
	"login.json":           {},
	"itemBundleItems.json": {},
	"itemDefinitions.json": {},
}

var Config = make(map[string]i.A)
var KeyedFiles = make(map[string]*orderedmap.OrderedMap)
var CloudFiles models.CloudFilesIndexMap
var Cloud = make(map[string][]byte)

func Initialize() {
	initializeCloud()
	initializeConfig()
}

func initializeConfig() {
	dirEntries, _ := static.Config.ReadDir("config")
	for _, entry := range dirEntries {
		data, err := fs.ReadFile(static.Config, "config/"+entry.Name())
		if err != nil {
			continue
		}
		switch entry.Name() {
		case "cloudfilesIndex.json":
			err = json.Unmarshal(data, &CloudFiles)
			for i, fileInfo := range CloudFiles {
				file := Cloud[i]
				fileInfo.Length = len(file)
				hash := md5.Sum(file)
				hashSlice := hash[:]
				fileInfo.Checksum = base64.StdEncoding.EncodeToString(hashSlice)
				CloudFiles[i] = fileInfo
			}
			break
		default:
			name := entry.Name()
			if _, keyed := keyedFilenames[name]; keyed {
				var result orderedmap.OrderedMap
				err = json.Unmarshal(data, &result)
				if err == nil {
					KeyedFiles[name] = &result
				}
			} else {
				var result i.A
				err = json.Unmarshal(data, &result)
				if err == nil {
					Config[name] = result
				}
			}
		}
	}
}

func initializeCloud() {
	dirEntries, _ := static.Cloud.ReadDir("cloud")
	for _, entry := range dirEntries {
		data, err := fs.ReadFile(static.Cloud, "cloud/"+entry.Name())
		if err == nil {
			Cloud[entry.Name()] = data
		}
	}
}

func ReturnSignedAsset(name string, c *gin.Context, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = KeyedFiles[name]
		rawSignature, _ := response.(*orderedmap.OrderedMap).Get("dataSignature")
		serverSignature = rawSignature.(string)
	} else {
		response = Config[name]
		arrayResponse := response.(i.A)
		serverSignature = arrayResponse[len(arrayResponse)-1].(string)
	}
	if c.Query("signature") != serverSignature {
		c.JSON(http.StatusOK, response)
		return
	}
	if keyedResponse {
		c.JSON(http.StatusOK, gin.H{"result": 0, "dataSignature": serverSignature})
	} else {
		emptyArrays := make(i.A, len(response.(i.A))-2)
		ret := i.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		c.JSON(http.StatusOK, ret)
	}
}
