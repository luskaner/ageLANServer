package asset

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
