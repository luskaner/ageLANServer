package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Buffer struct {
	f        *os.File
	buff     *bufio.Writer
	buffLock sync.Mutex
	enc      *json.Encoder
}

func (b *Buffer) Close() error {
	b.buffLock.Lock()
	defer b.buffLock.Unlock()
	if err := b.buff.Flush(); err != nil {
		return err
	}
	return b.f.Close()
}

func (b *Buffer) Log(value any) {
	if b == nil {
		return
	}
	b.buffLock.Lock()
	defer b.buffLock.Unlock()
	_ = b.enc.Encode(value)
}

func NewBuffer(f *os.File) *Buffer {
	buff := bufio.NewWriter(f)
	CommBuffer = &Buffer{
		f:    f,
		buff: buff,
		enc:  json.NewEncoder(buff),
	}
	return CommBuffer
}

var CommBuffer *Buffer

func Uptime(startTime *time.Time) time.Duration {
	if startTime == nil {
		startTime = new(time.Now())
	}
	return startTime.Sub(StartTime)
}
