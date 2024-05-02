package cloud

import (
	"embed"
	"io/fs"
)

//go:embed files/*
var content embed.FS
var Files = make(map[string][]byte)

func Initialize() {
	dirEntries, _ := content.ReadDir("files")
	for _, entry := range dirEntries {
		data, err := fs.ReadFile(content, "files/"+entry.Name())
		if err == nil {
			Files[entry.Name()] = data
		}
	}
}
