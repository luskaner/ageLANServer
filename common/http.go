package common

import "fmt"

func UserAgent() string {
	return fmt.Sprintf("%s/%s", Name, "1.0")
}
