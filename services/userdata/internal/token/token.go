// Package token verifies the HS256 JWT access tokens issued by the auth service
// (A-08). It is the read/verify counterpart to that service's issuer: same
// algorithm, same claim names (sub, iss, user_id, email, exp). Kong validates the
// token at the edge (A-09) and the BFF forwards the Authorization header, so this
// service re-verifies it to learn the authenticated user_id.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidToken is returned for any malformed, mis-signed, or wrong-issuer
// token. ErrExpired is returned specifically when the signature is valid but the
// token has expired.
var (
	ErrInvalidToken = errors.New("token: invalid")
	ErrExpired      = errors.New("token: expired")
)

// Claims is the verified subset of a token the service acts on.
type Claims struct {
	UserID string
	Email  string
}

// Verifier checks token signatures and claims against a shared secret + issuer.
type Verifier struct {
	secret []byte
	issuer string
	leeway time.Duration
	now    func() time.Time
}

// NewVerifier constructs a Verifier. issuer is the required "iss" claim; pass ""
// to skip the issuer check. leeway tolerates clock skew on expiry.
func NewVerifier(secret []byte, issuer string, leeway time.Duration) *Verifier {
	return &Verifier{secret: secret, issuer: issuer, leeway: leeway, now: time.Now}
}

type rawClaims struct {
	Sub    string `json:"sub"`
	Iss    string `json:"iss"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Verify checks the token's structure, signature, expiry, and issuer, and
// returns its claims. It returns ErrExpired for a valid-but-expired token and
// ErrInvalidToken for everything else.
func (v *Verifier) Verify(tokenString string) (Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	var hdr header
	if err := decodeSegment(parts[0], &hdr); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if hdr.Alg != "HS256" {
		return Claims{}, ErrInvalidToken
	}

	// Verify the signature over "header.payload" before trusting any claim.
	signing := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(signing))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(want, got) {
		return Claims{}, ErrInvalidToken
	}

	var claims rawClaims
	if err := decodeSegment(parts[1], &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	if v.issuer != "" && claims.Iss != v.issuer {
		return Claims{}, ErrInvalidToken
	}
	if claims.Exp != 0 && v.now().UTC().After(time.Unix(claims.Exp, 0).Add(v.leeway)) {
		return Claims{}, ErrExpired
	}

	userID := claims.UserID
	if userID == "" {
		userID = claims.Sub
	}
	if userID == "" {
		return Claims{}, ErrInvalidToken
	}
	return Claims{UserID: userID, Email: claims.Email}, nil
}

func decodeSegment(seg string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}
