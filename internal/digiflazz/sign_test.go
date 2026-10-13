package digiflazz

import "testing"

func TestExpectedSignDeterministic(t *testing.T) {
	a := ExpectedSign("user", "secret", "ref1")
	b := ExpectedSign("user", "secret", "ref1")
	if a != b || len(a) != 32 {
		t.Fatalf("unexpected sign: %s", a)
	}
}

func TestValidSign(t *testing.T) {
	sign := ExpectedSign("username", "apiKey", "some1d")
	if !ValidSign("username", "apiKey", "some1d", sign) {
		t.Fatal("expected valid sign")
	}
	if ValidSign("username", "apiKey", "some1d", "wrong") {
		t.Fatal("expected invalid sign")
	}
}
