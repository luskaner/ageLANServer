package models

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

type CloudfilesIndex struct {
	Id       int    `json:"id"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Version  string `json:"version"`
	ETag     string `json:"etag"`
	Created  string `json:"created"`
	Length   int
	Checksum string
}

type CloudFilesIndexMap map[string]CloudfilesIndex

func (m *CloudFilesIndexMap) GetByKey(key string) (string, *CloudfilesIndex, bool) {
	for filename, file := range *m {
		if file.Key == key {
			return filename, &file, true
		}
	}
	return "", nil, false
}

func (m *CloudFilesIndexMap) ReadFile(baseFolder string, name string) ([]byte, error) {
	_, ok := (*m)[name]
	if !ok {
		return nil, os.ErrInvalid
	}
	return os.ReadFile(filepath.Join(baseFolder, name))
}

func Build(configFolder string, baseFolder string) *CloudFilesIndexMap {
	data, err := os.ReadFile(filepath.Join(configFolder, "cloudfilesIndex.json"))
	if err != nil {
		panic(err)
	}
	var index CloudFilesIndexMap
	_ = json.Unmarshal(data, &index)
	for i, fileInfo := range index {
		data, err := index.ReadFile(baseFolder, i)
		if err != nil {
			panic(err)
		}
		fileInfo.Length = len(data)
		hash := md5.Sum(data)
		hashSlice := hash[:]
		fileInfo.Checksum = base64.StdEncoding.EncodeToString(hashSlice)
		index[i] = fileInfo
	}
	return &index
}
