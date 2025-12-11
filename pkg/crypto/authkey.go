package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

// CalcAuthKey is used for check whether key sent by client when login is correct or not.
func CalcAuthKey(key string, timestamp int64) string {
	m := md5.New()
	m.Write([]byte(key))
	m.Write([]byte(strconv.FormatInt(timestamp, 10)))
	res := m.Sum(nil)
	return hex.EncodeToString(res)
}
