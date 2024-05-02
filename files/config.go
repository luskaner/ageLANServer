package files

import (
	i "aoe2DELanServer/internal"
	"aoe2DELanServer/models"
	"aoe2DELanServer/static"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"github.com/wk8/go-ordered-map/v2"
	"io/fs"
	"net/http"
)

var keyedFilenames = map[string]struct{}{
	"login.json":           {},
	"itemBundleItems.json": {},
	"itemDefinitions.json": {},
}

var Config = make(map[string]i.A)
var KeyedFiles = make(map[string]*orderedmap.OrderedMap[string, any])
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
				var result = orderedmap.New[string, any]()
				err = json.Unmarshal(data, result)
				if err == nil {
					KeyedFiles[name] = result
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

func ReturnSignedAsset(name string, w *http.ResponseWriter, r *http.Request, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = KeyedFiles[name]
		rawSignature, _ := response.(*orderedmap.OrderedMap[string, any]).Get("dataSignature")
		serverSignature = rawSignature.(string)
	} else {
		response = Config[name]
		arrayResponse := response.(i.A)
		serverSignature = arrayResponse[len(arrayResponse)-1].(string)
	}
	if r.URL.Query().Get("signature") != serverSignature {
		i.JSON(w, response)
		return
	}
	if keyedResponse {
		i.JSON(w, i.H{"result": 0, "dataSignature": serverSignature})
	} else {
		emptyArrays := make(i.A, len(response.(i.A))-2)
		ret := i.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		i.JSON(w, ret)
	}
}
