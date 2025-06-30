package models

import (
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const resourceFolder = "resources"

var configFolder = filepath.Join(resourceFolder, "config")
var responsesFolder = filepath.Join(resourceFolder, "responses")
var CloudFolder = filepath.Join(responsesFolder, "cloud")

type file struct {
	path string
}

type arrayFile struct {
	file
	count int
}

type MainResources struct {
	keyedFilenames          mapset.Set[string]
	signedNonKeyedFilenames mapset.Set[string]
	ChatChannels            map[string]MainChatChannel
	LoginData               []i.A
	arrayFiles              map[string]*arrayFile
	keyedFiles              map[string]*file
	nameToSignature         map[string]string
	CloudFiles              CloudFiles
}

func (r *MainResources) Initialize(game common.GameTitle, keyedFilenames mapset.Set[string], signedNonKeyedFilenames mapset.Set[string]) {
	r.arrayFiles = make(map[string]*arrayFile)
	r.keyedFiles = make(map[string]*file)
	r.nameToSignature = make(map[string]string)
	r.keyedFilenames = keyedFilenames
	r.signedNonKeyedFilenames = signedNonKeyedFilenames
	r.initializeLogin(game)
	r.initializeChatChannels(game)
	r.initializeResponses(game)
	r.initializeCloud(game)
}

func (r *MainResources) initializeChatChannels(game common.GameTitle) {
	data, err := os.ReadFile(filepath.Join(configFolder, string(game), "chatChannels.json"))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &r.ChatChannels)
	if err != nil {
		panic(err)
	}
}

func (r *MainResources) initializeLogin(game common.GameTitle) {
	data, err := os.ReadFile(filepath.Join(configFolder, string(game), "login.json"))
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	for j := 0; j < len(matches)-1; j += 2 {
		r.LoginData = append(r.LoginData, i.A{matches[j][1], matches[j+1][1]})
	}
}

func (r *MainResources) initializeResponses(game common.GameTitle) {
	dirEntries, _ := os.ReadDir(filepath.Join(responsesFolder, string(game)))
	for _, entry := range dirEntries {
		fullPath := filepath.Join(responsesFolder, string(game), entry.Name())
		data, err := os.ReadFile(fullPath)
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
			if len(matches) == 2 {
				r.keyedFiles[name] = &file{path: fullPath}
				r.nameToSignature[name] = strings.Clone(matches[1])
			}
		} else {
			var result i.A
			err = json.Unmarshal(data, &result)
			if err == nil {
				r.arrayFiles[name] = &arrayFile{
					file: file{
						path: fullPath,
					},
					count: len(result),
				}
				if r.signedNonKeyedFilenames.ContainsOne(name) {
					r.nameToSignature[name] = strings.Clone(result[r.arrayFiles[name].count-1].(string))
				}
			}
		}
	}
}

func (r *MainResources) initializeCloud(gameTitle common.GameTitle) {
	cloudfiles := BuildCloudfilesIndex(filepath.Join(configFolder, string(gameTitle)), filepath.Join(CloudFolder, string(gameTitle)))
	if cloudfiles != nil {
		r.CloudFiles = *cloudfiles
	}
}

func (r *MainResources) ReturnArrayFile(name string, w *http.ResponseWriter) {
	data, _ := os.ReadFile(r.arrayFiles[name].path)
	i.RawJSON(w, data)
}

func (r *MainResources) ReturnSignedAsset(name string, w *http.ResponseWriter, req *http.Request) {
	serverSignature := r.nameToSignature[name]
	var f *file
	var count int
	var keyedResponse bool
	if f, keyedResponse = r.keyedFiles[name]; !keyedResponse {
		arrFile := r.arrayFiles[name]
		f = &arrFile.file
		count = arrFile.count
	}
	if req.URL.Query().Get("signature") != serverSignature {
		response, _ := os.ReadFile(f.path)
		i.RawJSON(w, response)
		return
	}
	if keyedResponse {
		i.RawJSON(w, []byte(fmt.Sprintf(`{"result":0,"dataSignature":"%s"}`, serverSignature)))
	} else {
		emptyArrays := make(i.A, count-2)
		ret := i.A{0}
		ret = append(ret, emptyArrays...)
		ret = append(ret, serverSignature)
		i.JSON(w, ret)
	}
}
