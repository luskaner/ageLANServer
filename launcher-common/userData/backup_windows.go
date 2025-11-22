package userData

import "os"

func basePath(_ string) string {
	return os.Getenv("USERPROFILE")
}
