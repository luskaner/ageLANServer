package router

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/luskaner/ageLANServer/server/internal/logger"
)

func formatDurationWithDays(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %02d:%02d:%02d",
			days, hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d:%02d",
		hours, minutes, seconds)
}

func logFormatter(writer io.Writer, params handlers.LogFormatterParams) {
	host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
	if err != nil {
		host = params.Request.RemoteAddr
	}
	uri := params.Request.RequestURI
	if params.Request.ProtoMajor == 2 && params.Request.Method == "CONNECT" {
		uri = params.Request.Host
	}
	if uri == "" {
		uri = params.Request.URL.RequestURI()
	}
	t := params.TimeStamp.UTC().Sub(logger.StartTime)
	var buf strings.Builder
	buf.WriteString(host)
	buf.WriteString(" - ")
	buf.WriteString(formatDurationWithDays(t))
	buf.WriteString(" - ")
	buf.WriteString(params.Request.Method)
	buf.WriteString(" ")
	buf.WriteString(uri)
	buf.WriteString(" - ")
	buf.WriteString(strconv.Itoa(params.StatusCode))
	buf.WriteString(" ")
	buf.WriteString(strconv.Itoa(params.Size))
	buf.WriteString("\n")
	_, _ = writer.Write([]byte(buf.String()))
}
