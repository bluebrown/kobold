package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

const SEPERATOR = "\u2342"

type SliceText []string

func (s *SliceText) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("dbutil: cannot convert %T to SliceText", value)
	}

	out := strings.Split(str, SEPERATOR)

	*s = out

	return nil
}

func (s SliceText) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "", nil
	}

	return strings.Join(s, SEPERATOR), nil
}

type JsonArray []string

func (s *JsonArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return fmt.Errorf("dbutil: cannot convert %T to JsonArray", value)
	}
	return json.Unmarshal(b, &s)
}

func (s JsonArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "", nil
	}
	return json.Marshal(s)
}
