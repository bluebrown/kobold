package dbutil

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	sqlite3 "modernc.org/sqlite"
)

func sum(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func TaskFingerPrint(ids []string) string {
	b := []byte(strings.Join(ids, ","))
	return sum(b)
}

func MustMakeUUID() {
	sqlite3.MustRegisterDeterministicScalarFunction(
		"uuid",
		0,
		func(ctx *sqlite3.FunctionContext, args []driver.Value) (driver.Value, error) {
			r, err := uuid.NewRandom()
			if err != nil {
				return nil, err
			}
			return r.String(), nil
		},
	)
}

func MustMakeSha1() {
	sqlite3.MustRegisterDeterministicScalarFunction(
		"sha1",
		1,
		func(ctx *sqlite3.FunctionContext, args []driver.Value) (driver.Value, error) {
			var b []byte
			switch v := args[0].(type) {
			case []byte:
				b = v
			case string:
				b = []byte(v)
			default:
				return nil, fmt.Errorf("invalid type: %T", v)
			}
			return sum(b), nil
		},
	)
}
