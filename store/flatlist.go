package store

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// this is a special json array. When scanning, it will flatten the input up to
// 1 level deep. however, when calling value, it will only marshal a flat list.
// so the converion is lossy, but it fits our use case perfectly. We use it to
// either store flat lists on actual tables or to retrieve potentially nested
// lists from accumulated views.
type FlatList []string

func (s FlatList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "", nil
	}
	return json.Marshal(s)
}

func (s *FlatList) Scan(value interface{}) error {
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
		return fmt.Errorf("store: cannot convert %T to FlatList", value)
	}

	if len(b) == 0 {
		return nil
	}

	var iterm []any
	if err := json.Unmarshal(b, &iterm); err != nil {
		return fmt.Errorf("unmarshal intermeditate: %w", err)
	}

	for _, item := range iterm {
		switch v := item.(type) {
		case string:
			*s = append(*s, v)
		case []interface{}:
			for _, vv := range v {
				*s = append(*s, vv.(string))
			}
		default:
			return fmt.Errorf("store: cannot convert %T to FlatList", value)
		}
	}

	return nil
}
