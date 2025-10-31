package commonLogger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/luskaner/ageLANServer/common"
)

var logger *log.Logger
var file *os.File
var FileLogger *Root
var buf bufferWrapper

type Root struct {
	root string
}

type bufferWrapper struct {
	buffer bytes.Buffer
}

func (b *bufferWrapper) Write(p []byte) (n int, err error) {
	n, err = b.buffer.Write(p)
	return n, err
}

func NewOwnFileLogger(name string, root string, gameId string, finalRoot bool) error {
	fileLogger, f, err := NewFileLogger(name, root, gameId, finalRoot)
	if err != nil {
		return err
	}
	FileLogger = fileLogger
	file = f
	return nil
}

func NewFileLogger(name string, root string, gameId string, finalRoot bool) (fileLogger *Root, f *os.File, err error) {
	err, fileLogger = NewFile(root, gameId, finalRoot)
	if err != nil {
		return
	}
	f, err = fileLogger.Open(name)
	return
}

func Buffer(name string, root string, gameId string, finalRoot bool, fn func(writer io.Writer)) error {
	err, fileLogger := NewFile(root, gameId, finalRoot)
	if err != nil {
		return err
	}
	return fileLogger.Buffer(name, fn)
}

func Initialize(writer io.Writer) {
	var flags int
	if writer == nil {
		writer = &buf
	}
	if writer != os.Stdout || !common.Interactive() {
		flags = log.Lmicroseconds | log.Ltime | log.LUTC | log.Lmsgprefix
	}
	logger = log.New(writer, "", flags)
}

func Prefix(name string) {
	logger.SetPrefix("|" + strings.ToUpper(name) + "| ")
}

func CloseFileLog() {
	if file != nil {
		if _, err := file.Write(buf.buffer.Bytes()); err != nil {
			return
		}
		_ = file.Sync()
		_ = file.Close()
	}
}

func PrefixPrintf(name string, format string, a ...any) {
	if logger != nil {
		Prefix(name)
		Printf(format, a...)
	}
}

func Printf(format string, a ...any) {
	if logger != nil {
		logger.Printf(format, a...)
	}
}

func PrefixPrintln(name string, a ...any) {
	if logger != nil {
		Prefix(name)
		Println(a...)
	}
}

func Println(a ...any) {
	if logger != nil {
		logger.Println(a...)
	}
}

func logRootPrefix(root string) string {
	return filepath.Join(root, "logs")
}

func LogRootDate(root string) string {
	return filepath.Join(logRootPrefix(root), date())
}

func date() string {
	t := time.Now().UTC()
	return fmt.Sprintf("%d-%02d-%02dT%02d-%02d-%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func NewFile(root string, gameId string, finalRoot bool) (err error, l *Root) {
	folder := root
	if !finalRoot {
		folder = filepath.Join(logRootPrefix(root), gameId, date())
	}
	folder, err = filepath.Abs(folder)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(folder, 0755)
	if err != nil {
		return err, nil
	}
	return nil, &Root{root: folder}
}

func (l *Root) Open(name string) (f *os.File, err error) {
	if l == nil {
		return
	}
	return os.OpenFile(
		filepath.Join(l.root, name+".txt"),
		os.O_CREATE|os.O_WRONLY,
		0666,
	)
}

func (l *Root) Folder() string {
	if l == nil {
		return ""
	}
	return l.root
}

func (l *Root) Buffer(name string, fn func(writer io.Writer)) error {
	if l == nil {
		fn(nil)
		return nil
	}
	var buff bytes.Buffer
	fn(&buff)
	if buff.Len() > 0 {
		f, err := l.Open(name)
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		_, err = io.Copy(f, &buff)
		if err != nil {
			return err
		}
	}
	return nil
}
