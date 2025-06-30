package userData

import (
	"github.com/luskaner/ageLANServer/common"
	"os"
)

func basePath(_ common.GameTitle) string {
	return os.Getenv("USERPROFILE")
}
