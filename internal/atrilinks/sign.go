package atrilinks

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/zeebo/blake3"
)

// Sign menghitung tanda tangan BLAKE3 (hex lowercase, 64 karakter) untuk metode Atrilinks.
// Rumus: combined = reqid|kodeproduk|tujuan:userid:password:pin:Atrilinks
func Sign(reqid, kodeproduk, tujuan, userid, password, pin string) string {
	combined := fmt.Sprintf("%s|%s|%s:%s:%s:%s:Atrilinks",
		reqid, kodeproduk, tujuan, userid, password, pin,
	)
	sum := blake3.Sum256([]byte(combined))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}
