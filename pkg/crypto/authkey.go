package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// CalcAuthKey is used for check whether key sent by client when login is correct or not.
func CalcAuthKey(key string, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(key))
	h.Write([]byte(strconv.FormatInt(timestamp, 10)))
	res := h.Sum(nil)
	return hex.EncodeToString(res)
}
