package userData

import (
	"path/filepath"
	"strings"

	"github.com/luskaner/ageLANServer/common"
)

const backupSuffix = ".bak"
const serverSuffix = ".lan"
const prefix = "Games"

const (
	TypeActive = iota
	TypeServer
	TypeBackup
)

var suffixToType = map[string]int{
	backupSuffix: TypeBackup,
	serverSuffix: TypeServer,
}

type Data struct {
	Type int
	Path string
}

func Path(gameId string) string {
	var s string
	addPrefix := true
	switch gameId {
	case common.GameAoE1:
		s = `Age of Empires DE`
	case common.GameAoE2:
		s = `Age of Empires 2 DE`
	case common.GameAoE3:
		s = `Age of Empires 3 DE`
	case common.GameAoE4:
		addPrefix = false
		s = `Age of Empires IV`
	case common.GameAoM:
		s = `Age of Mythology Retold`
	}
	path := basePath(gameId)
	if addPrefix {
		path = filepath.Join(path, prefix)
	}
	return filepath.Join(path, s)
}

func typ(path string) (typ int, ext string) {
	ext = filepath.Ext(path)
	var ok bool
	if typ, ok = suffixToType[ext]; !ok {
		typ = TypeActive
	}
	return
}

func suffix(typ int) string {
	for currentSuffix, currentType := range suffixToType {
		if currentType == typ {
			return currentSuffix
		}
	}
	return ""
}

func TransformPath(path string, srcType int, dstType int) (ok bool, transformedPath string) {
	t, ext := typ(path)
	if t != srcType {
		ok = false
		return
	}
	if t == dstType {
		ok = true
		transformedPath = path
		return
	}
	ok = true
	transformedPath = strings.TrimSuffix(path, ext) + suffix(dstType)
	return
}
