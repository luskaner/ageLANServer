package userData

import (
	"maps"
	"os"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

func metadataFolder(gameId string) string {
	var p string
	switch gameId {
	case common.GameAoE2:
		p = "metadata"
	case common.GameAoE3:
		p = filepath.Join("Common", "RLink")
	case common.GameAoM:
		p = filepath.Join("temp", "RLink")
	}
	return p
}

func Metadatas(gameId string) (err error, metadatas mapset.Set[Data]) {
	p := filepath.Join(Path(gameId), metadataFolder(gameId))
	metadatas = mapset.NewThreadUnsafeSet[Data]()
	if p != "" {
		allSuffixes := maps.Clone(suffixToType)
		allSuffixes[""] = TypeActive
		for _, t := range allSuffixes {
			if ok, transformedPath := TransformPath(p, TypeActive, t); ok {
				if f, localErr := os.Stat(transformedPath); localErr == nil && f.IsDir() {
					metadatas.Add(Data{t, transformedPath})
				}
			}
		}
	}
	return
}
