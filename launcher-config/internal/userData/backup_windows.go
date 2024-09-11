package userData

import "os"

func basePath() string {
	return os.Getenv("USERPROFILE")
}
