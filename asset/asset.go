package asset

import (
	"aoe2DELanServer/asset/cloud"
	"aoe2DELanServer/j"
	"crypto/md5"
	"embed"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

import "github.com/iancoleman/orderedmap"

//go:embed files/*.json
var content embed.FS

var keyedFilenames = map[string]struct{}{
	"configuration.json":     {},
	"item_bundle_items.json": {},
	"item_definitions.json":  {},
}

var Files = make(map[string]j.A)
var KeyedFiles = make(map[string]*orderedmap.OrderedMap)
var CloudFiles CloudFilesIndexMap

// var Achievements j.A

func Initialize() {
	dirEntries, _ := content.ReadDir("files")
	for _, entry := range dirEntries {
		data, err := fs.ReadFile(content, "files/"+entry.Name())
		if err != nil {
			continue
		}
		switch entry.Name() {
		case "cloudfiles_index.json":
			err = json.Unmarshal(data, &CloudFiles)
			for i, fileInfo := range CloudFiles {
				file := cloud.Files[i]
				fileInfo.Length = len(file)
				hash := md5.Sum(file)
				hashSlice := hash[:]
				fileInfo.Checksum = base64.StdEncoding.EncodeToString(hashSlice)
				CloudFiles[i] = fileInfo
			}
			break
		case "achievements.json":
			var result j.A
			err = json.Unmarshal(data, &result)
			if err == nil {
				Files[entry.Name()] = result
			}
			/*achievementsJson := result[1].([]interface{})
			Achievements = make(j.A, len(achievementsJson))
			for i, achievement := range achievementsJson {
				achievementInfo := achievement.([]interface{})
				Achievements[i] = j.A{uint32(achievementInfo[0].(float64)), uint64(achievementInfo[5].(float64))}
			}*/
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
				var result j.A
				err = json.Unmarshal(data, &result)
				if err == nil {
					Files[name] = result
				}
			}
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
		response = Files[name]
		arrayResponse := response.(j.A)
		serverSignature = arrayResponse[len(arrayResponse)-1].(string)
	}
	if c.Query("signature") != serverSignature {
		c.JSON(http.StatusOK, response)
		return
	}
	if keyedResponse {
		c.JSON(http.StatusOK, gin.H{"result": 0, "dataSignature": serverSignature})
	} else {
		emptyArrays := make(j.A, len(response.(j.A))-2)
		ret := j.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		c.JSON(http.StatusOK, ret)
	}
}
