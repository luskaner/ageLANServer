package logger

import (
	"fmt"
	"os"
	"time"

	commonLogger "github.com/luskaner/ageLANServer/common/logger"
)

var StartTime time.Time

func init() {
	StartTime = time.Now().UTC()
}

func OpenMainFileLog(root string, logEnabled bool) error {
	if logEnabled {
		err := commonLogger.NewOwnFileLogger("server", root, "", true)
		if err != nil {
			return err
		}
	}
	return nil
}

func PrintFile(name string, path string) {
	if commonLogger.FileLogger != nil && path != "" {
		data, _ := os.ReadFile(path)
		commonLogger.PrefixPrintln(name, string(data))
	}
}

func Printf(format string, a ...any) {
	commonLogger.PrefixPrintf("main", format, a...)
	fmt.Printf(format, a...)
}

func Println(a ...any) {
	commonLogger.PrefixPrintln("main", a...)
	fmt.Println(a...)
}
