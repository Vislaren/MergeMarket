package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// mint signs a token the same way the auth service (A-08) does, so these tests
// exercise the real cross-service contract.
func mint(secret []byte, claims map[string]any) string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	hb, _ := json.Marshal(header)
	cb, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(cb)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validClaims() map[string]any {
	return map[string]any{
		"sub":     "user-123",
		"iss":     "mergemarket-auth",
		"user_id": "user-123",
		"email":   "a@b.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
}

func TestVerifyValid(t *testing.T) {
	secret := []byte("secret")
	v := NewVerifier(secret, "mergemarket-auth", 30*time.Second)
	claims, err := v.Verify(mint(secret, validClaims()))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != "user-123" || claims.Email != "a@b.com" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestVerifyFallsBackToSub(t *testing.T) {
	secret := []byte("secret")
	c := validClaims()
	delete(c, "user_id")
	v := NewVerifier(secret, "mergemarket-auth", 0)
	claims, err := v.Verify(mint(secret, c))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want fallback to sub", claims.UserID)
	}
}

func TestVerifyExpired(t *testing.T) {
	secret := []byte("secret")
	c := validClaims()
	c["exp"] = time.Now().Add(-time.Hour).Unix()
	v := NewVerifier(secret, "mergemarket-auth", 0)
	_, err := v.Verify(mint(secret, c))
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("error = %v, want ErrExpired", err)
	}
}

func TestVerifyWrongSecret(t *testing.T) {
	v := NewVerifier([]byte("right"), "mergemarket-auth", 0)
	_, err := v.Verify(mint([]byte("wrong"), validClaims()))
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyWrongIssuer(t *testing.T) {
	secret := []byte("secret")
	v := NewVerifier(secret, "mergemarket-auth", 0)
	c := validClaims()
	c["iss"] = "evil"
	_, err := v.Verify(mint(secret, c))
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyTampered(t *testing.T) {
	secret := []byte("secret")
	v := NewVerifier(secret, "mergemarket-auth", 0)
	tok := mint(secret, validClaims())
	// Flip a character in the payload segment.
	parts := strings.Split(tok, ".")
	parts[1] = parts[1][:len(parts[1])-1] + "A"
	_, err := v.Verify(strings.Join(parts, "."))
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyMalformed(t *testing.T) {
	v := NewVerifier([]byte("secret"), "", 0)
	for _, bad := range []string{"", "abc", "a.b", "a.b.c.d"} {
		if _, err := v.Verify(bad); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Verify(%q) error = %v, want ErrInvalidToken", bad, err)
		}
	}
}
