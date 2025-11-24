package playfab

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

type target struct {
	CdnBranch string
}

type rule struct {
	target `json:"Target"`
}

type cdnBundle struct {
	Id                  string
	CdnBranch           string
	RequiresGameVersion int
}

type cdnPathConfig struct {
	Rules      []rule
	CdnBundles []cdnBundle
}

var BaseDir = path.Join(models.ResponsesFolder, common.GameAoM, "playfab")
var StaticConfig string

const StaticSuffix = "/static"
const branch = "public/production"

/* Ids:
* 1: c8c9727eb975e7aba1f949beaa6153e7e7ccb415
* 2: 1edf70cddb1a814c20cc368c3770a60ab2657358
 */

func init() {
	c := cdnPathConfig{
		Rules: []rule{
			{
				target{CdnBranch: branch},
			},
		},
	}
	_ = filepath.Walk(BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			id, err := strconv.Atoi(info.Name())
			if err != nil {
				return nil
			}
			c.CdnBundles = append(
				c.CdnBundles,
				cdnBundle{
					info.Name(),
					branch,
					id,
				},
			)
		}
		return nil
	})
	slices.SortFunc(c.CdnBundles, func(a, b cdnBundle) int {
		if a.RequiresGameVersion > b.RequiresGameVersion {
			return -1
		}
		if a.RequiresGameVersion < b.RequiresGameVersion {
			return 1
		}
		return 0
	})
	str, err := json.Marshal(c)
	if err == nil {
		StaticConfig = string(str)
	}
}
