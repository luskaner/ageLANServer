package playfab

import (
	"encoding/json"
	"time"
)

const CustomTimeFormat = "2006-01-02T15:04:05.000Z"

type CustomTime struct {
	time.Time
	Format string
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formatted := ct.Time.UTC().Format(CustomTimeFormat)
	return json.Marshal(formatted)
}

type ValueLike interface {
	Prepare() error
}

type Value[T any] struct {
	Val         *T `json:"-"`
	Value       string
	LastUpdated CustomTime
	Permission  string
}

func (v *Value[T]) Prepare() error {
	v.LastUpdated = CustomTime{time.Now(), "2006-01-02T15:04:05.000Z"}
	v.Permission = "Private"
	if v.Val != nil {
		if val, err := json.Marshal(v.Val); err == nil {
			v.Value = string(val)
			return nil
		} else {
			return err
		}
	}
	return nil
}
