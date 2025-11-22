package logger

import (
	"fmt"
	"time"

	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/viper"
)

var StartTime time.Time

func OpenMainFileLog(root string) error {
	if viper.GetBool("Config.Log") {
		err := commonLogger.NewOwnFileLogger("server", root, "", true)
		if err != nil {
			return err
		}
	}
	return nil
}

func Printf(format string, a ...any) {
	commonLogger.PrefixPrintf("main", format, a...)
	fmt.Printf(format, a...)
}

func Println(a ...any) {
	commonLogger.PrefixPrintln("main", a...)
	fmt.Println(a...)
}
