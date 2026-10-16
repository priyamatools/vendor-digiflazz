package atrilinks

import (
	"testing"
)

// TestSignFormat memastikan output hex 64 karakter lowercase.
func TestSignFormat(t *testing.T) {
	s := Sign("TRX001", "xl10", "08123456789", "LS001", "secretpass", "1234")
	if len(s) != 64 {
		t.Fatalf("panjang sign harus 64, got %d: %s", len(s), s)
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Fatalf("sign harus hex lowercase: %s", s)
		}
	}
	a := Sign("a", "b", "c", "d", "e", "f")
	b := Sign("a", "b", "c", "d", "e", "f")
	if a != b {
		t.Fatal("sign harus deterministik")
	}
}
