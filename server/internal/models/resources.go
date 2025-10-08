package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	i "github.com/luskaner/ageLANServer/server/internal"
)

const resourceFolder = "resources"

var configFolder = filepath.Join(resourceFolder, "config")
var responsesFolder = filepath.Join(resourceFolder, "responses")
var CloudFolder = filepath.Join(responsesFolder, "cloud")

type MainResources struct {
	keyedFilenames  mapset.Set[string]
	ChatChannels    map[string]MainChatChannel
	LoginData       []i.A
	ArrayFiles      map[string]i.A
	KeyedFiles      map[string][]byte
	nameToSignature map[string]string
	CloudFiles      CloudFiles
}

func (r *MainResources) Initialize(gameId string, keyedFilenames mapset.Set[string]) {
	r.ArrayFiles = make(map[string]i.A)
	r.KeyedFiles = make(map[string][]byte)
	r.nameToSignature = make(map[string]string)
	r.keyedFilenames = keyedFilenames
	r.initializeLogin(gameId)
	r.initializeChatChannels(gameId)
	r.initializeResponses(gameId)
	r.initializeCloud(gameId)
}

func (r *MainResources) initializeChatChannels(gameId string) {
	data, err := os.ReadFile(filepath.Join(configFolder, gameId, "chatChannels.json"))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &r.ChatChannels)
	if err != nil {
		panic(err)
	}
}

func (r *MainResources) initializeLogin(gameId string) {
	data, err := os.ReadFile(filepath.Join(configFolder, gameId, "login.json"))
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	for j := 0; j < len(matches)-1; j += 2 {
		r.LoginData = append(r.LoginData, i.A{matches[j][1], matches[j+1][1]})
	}
}

func (r *MainResources) initializeResponses(gameId string) {
	dirEntries, _ := os.ReadDir(filepath.Join(responsesFolder, gameId))
	for _, entry := range dirEntries {
		data, err := os.ReadFile(filepath.Join(responsesFolder, gameId, entry.Name()))
		if err != nil {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if r.keyedFilenames.ContainsOne(name) {
			re := regexp.MustCompile(`"dataSignature"\s*:\s*"(.*?)"`)
			matches := re.FindStringSubmatch(string(data))
			if len(matches) == 1 {
				serverSignature := matches[1]
				r.KeyedFiles[name] = data
				r.nameToSignature[name] = serverSignature
			}
		} else {
			var result i.A
			err = json.Unmarshal(data, &result)
			if err == nil {
				r.ArrayFiles[name] = result
			}
		}
	}
}

func (r *MainResources) initializeCloud(gameId string) {
	cloudfiles := BuildCloudfilesIndex(filepath.Join(configFolder, gameId), filepath.Join(CloudFolder, gameId))
	if cloudfiles != nil {
		r.CloudFiles = *cloudfiles
	}
}

func (r *MainResources) ReturnSignedAsset(name string, w *http.ResponseWriter, req *http.Request, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = r.KeyedFiles[name]
		serverSignature = r.nameToSignature[name]
	} else {
		response = r.ArrayFiles[name]
		arrayResponse := response.(i.A)
		serverSignature = arrayResponse[len(arrayResponse)-1].(string)
	}
	if req.URL.Query().Get("signature") != serverSignature {
		if keyedResponse {
			i.RawJSON(w, response.([]byte))
		} else {
			i.JSON(w, response)
		}
		return
	}
	if keyedResponse {
		i.RawJSON(w, []byte(fmt.Sprintf(`{"result":0,"dataSignature":"%s"}`, serverSignature)))
	} else {
		emptyArrays := make(i.A, len(response.(i.A))-2)
		ret := i.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		i.JSON(w, ret)
	}
}
