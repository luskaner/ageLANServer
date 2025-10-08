package shared

import "time"

const dateFormat = "2006-01-02T15:04:05.000Z"

func FormatDate(date time.Time) string {
	return date.Format(dateFormat)
}
