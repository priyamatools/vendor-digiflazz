package digiflazz

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// ExpectedSign menghitung md5(username + apiKey + ref_id) untuk validasi request Digiflazz.
func ExpectedSign(username, apiKey, refID string) string {
	sum := md5.Sum([]byte(username + apiKey + refID))
	return hex.EncodeToString(sum[:])
}

// ValidSign membandingkan sign request dengan nilai yang diharapkan (constant-time).
func ValidSign(username, apiKey, refID, sign string) bool {
	expected := ExpectedSign(username, apiKey, refID)
	if len(sign) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sign), []byte(expected)) == 1
}

// PriceString mengubah harga int menjadi string untuk response Digiflazz.
func PriceString(price int64) string {
	return fmt.Sprintf("%d", price)
}
