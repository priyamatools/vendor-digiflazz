package digiflazz

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func ExpectedSign(username, apiKey, refID string) string {
	sum := md5.Sum([]byte(username + apiKey + refID))
	return hex.EncodeToString(sum[:])
}

func ValidSign(username, apiKey, refID, sign string) bool {
	expected := ExpectedSign(username, apiKey, refID)
	if len(sign) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sign), []byte(expected)) == 1
}

func PriceString(price int64) string {
	return fmt.Sprintf("%d", price)
}
