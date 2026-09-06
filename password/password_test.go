package password

import "testing"

func TestHashVerify(t *testing.T) {
	hash, err := HashWithCost("secret", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(hash, "secret") || Verify(hash, "wrong") {
		t.Fatal("password verification failed")
	}
}
