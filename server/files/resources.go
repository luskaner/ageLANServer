package files

import (
	"encoding/json"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"net/http"
	"os"
	"path/filepath"
	i "server/internal"
	"server/models"
	"strings"
)

var keyedFilenames = map[string]struct{}{
	"itemBundleItems.json": {},
	"itemDefinitions.json": {},
}

var Login []i.A
var ArrayFiles = make(map[string]i.A)
var KeyedFiles = make(map[string]*orderedmap.OrderedMap[string, any])
var CloudFiles models.CloudFilesIndexMap

func Initialize() {
	initializeLogin()
	initializeResponses()
	initializeCloud()
}

const resourceFolder = "resources"

var configFolder = filepath.Join(resourceFolder, "config")
var responsesFolder = filepath.Join(resourceFolder, "responses")
var CloudFolder = filepath.Join(responsesFolder, "cloud")

func initializeLogin() {
	data, err := os.ReadFile(filepath.Join(configFolder, "login.json"))
	if err != nil {
		panic(err)
	}
	var login = orderedmap.New[string, any]()
	err = json.Unmarshal(data, login)
	if err != nil {
		panic(err)
	}
	Login = make([]i.A, login.Len())
	j := 0
	for el := login.Oldest(); el != nil; el = el.Next() {
		Login[j] = i.A{el.Key, el.Value}
		j++
	}
}

func initializeResponses() {
	dirEntries, _ := os.ReadDir(responsesFolder)
	for _, entry := range dirEntries {
		data, err := os.ReadFile(filepath.Join(responsesFolder, entry.Name()))
		if err != nil {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
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
				ArrayFiles[name] = result
			}
		}
	}
}

func initializeCloud() {
	CloudFiles = *models.Build(configFolder, CloudFolder)
}

func ReadCloudFile(name string) ([]byte, error) {
	return CloudFiles.ReadFile(CloudFolder, name)
}

func ReturnSignedAsset(name string, w *http.ResponseWriter, r *http.Request, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = KeyedFiles[name]
		rawSignature, _ := response.(*orderedmap.OrderedMap[string, any]).Get("dataSignature")
		serverSignature = rawSignature.(string)
	} else {
		response = ArrayFiles[name]
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
