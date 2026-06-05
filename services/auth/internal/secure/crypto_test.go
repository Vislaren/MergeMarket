package secure

import "testing"

func TestEncryptDecryptString(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptString(key, "secret")
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}
	if encrypted == "secret" {
		t.Fatalf("ciphertext must not equal plaintext")
	}
	plain, err := DecryptString(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}
	if plain != "secret" {
		t.Fatalf("plain = %q, want secret", plain)
	}
}
