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
	"github.com/luskaner/ageLANServer/common/paths"
	i "github.com/luskaner/ageLANServer/server/internal"
)

var ResponsesFolder = filepath.Join(paths.ResourcesDir, "responses")
var userDataFolder = filepath.Join(paths.ResourcesDir, "userData")
var CloudFolder = filepath.Join(ResponsesFolder, "cloud")

type ResourcesOpts struct {
	KeyedFilenames mapset.Set[string]
}

type signature struct {
	Value string `json:"dataSignature"`
}

type Resources interface {
	Initialize(gameId string, opts *ResourcesOpts)
	ReturnSignedAsset(name string, w *http.ResponseWriter, req *http.Request, keyedResponse bool)
	LoginData() []i.A
	ChatChannels() map[string]*MainChatChannel
	ArrayFiles() map[string]i.A
	SignedAssets() map[string][]byte
	CloudFiles() CloudFiles
}

type MainResources struct {
	keyedFilenames  mapset.Set[string]
	chatChannels    map[string]*MainChatChannel
	loginData       []i.A
	arrayFiles      map[string]i.A
	keyedFiles      map[string][]byte
	nameToSignature map[string]string
	cloudFiles      CloudFiles
}

func (r *MainResources) Initialize(gameId string, opts *ResourcesOpts) {
	if opts == nil {
		opts = &ResourcesOpts{}
	}
	if opts.KeyedFilenames == nil {
		opts.KeyedFilenames = mapset.NewSet[string]("itemDefinitions.json")
	}
	r.arrayFiles = make(map[string]i.A)
	r.keyedFiles = make(map[string][]byte)
	r.nameToSignature = make(map[string]string)
	r.keyedFilenames = opts.KeyedFilenames
	r.initializeUserData(gameId)
	r.initializeLogin(gameId)
	r.initializeChatChannels(gameId)
	r.initializeResponses(gameId)
	r.initializeCloud(gameId)
}

func (r *MainResources) LoginData() []i.A {
	return r.loginData
}

func (r *MainResources) ChatChannels() map[string]*MainChatChannel {
	return r.chatChannels
}

func (r *MainResources) ArrayFiles() map[string]i.A {
	return r.arrayFiles
}

func (r *MainResources) SignedAssets() map[string][]byte {
	return r.keyedFiles
}

func (r *MainResources) CloudFiles() CloudFiles {
	return r.cloudFiles
}

func (r *MainResources) initializeChatChannels(gameId string) {
	data, err := os.ReadFile(filepath.Join(paths.ConfigsPath, gameId, "chatChannels.json"))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &r.chatChannels)
	if err != nil {
		panic(err)
	}
}

func (r *MainResources) initializeLogin(gameId string) {
	data, err := os.ReadFile(filepath.Join(paths.ConfigsPath, gameId, "login.json"))
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`"([^"]+)"\s*:\s*"([^"]*)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	for _, m := range matches {
		if len(m) == 3 {
			r.loginData = append(r.loginData, i.A{m[1], m[2]})
		}
	}
}

func (r *MainResources) initializeResponses(gameId string) {
	dirEntries, _ := os.ReadDir(filepath.Join(ResponsesFolder, gameId))
	for _, entry := range dirEntries {
		data, err := os.ReadFile(filepath.Join(ResponsesFolder, gameId, entry.Name()))
		if err != nil {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if r.keyedFilenames.ContainsOne(name) {
			var sig signature
			err = json.Unmarshal(data, &sig)
			if err == nil && sig.Value != "" {
				r.keyedFiles[name] = data
				r.nameToSignature[name] = sig.Value
			}
		} else {
			var result i.A
			err = json.Unmarshal(data, &result)
			if err == nil {
				r.arrayFiles[name] = result
			}
		}
	}
}

func (r *MainResources) initializeCloud(gameId string) {
	cloudfiles := BuildCloudfilesIndex(filepath.Join(paths.ConfigsPath, gameId), filepath.Join(CloudFolder, gameId))
	if cloudfiles != nil {
		r.cloudFiles = *cloudfiles
	}
}

func (r *MainResources) ReturnSignedAsset(name string, w *http.ResponseWriter, req *http.Request, keyedResponse bool) {
	var serverSignature string
	var response any
	if keyedResponse {
		response = r.keyedFiles[name]
		serverSignature = r.nameToSignature[name]
	} else {
		response = r.arrayFiles[name]
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

func (r *MainResources) initializeUserData(gameId string) {
	ensureFolder(filepath.Join(userDataFolder, gameId))
}

func ensureFolder(path string) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}
}

func UserDataPath(gameId string, steam bool, platformUserid string) string {
	var platform string
	if steam {
		platform = "STEAM"
	} else {
		platform = "XBOX"
	}
	folder := filepath.Join(userDataFolder, gameId)
	ensureFolder(folder)
	return filepath.Join(folder, fmt.Sprintf("%s_%s", platform, platformUserid)+".json")
}
