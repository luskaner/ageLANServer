package models

type CloudfilesIndex struct {
	Id       int    `schema:"id"`
	Key      string `schema:"key"`
	Type     string `schema:"type"`
	Version  string `schema:"version"`
	ETag     string `schema:"etag"`
	Created  string `schema:"created"`
	Length   int
	Checksum string
}

type CloudFilesIndexMap map[string]CloudfilesIndex
