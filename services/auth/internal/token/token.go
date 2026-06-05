// Package token issues signed JWT access tokens for authenticated users.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Issuer signs MergeMarket access tokens.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// Pair is the client-facing token bundle.
type Pair struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewIssuer constructs a JWT issuer with the configured HMAC secret and TTL.
func NewIssuer(secret []byte, ttl time.Duration) *Issuer {
	return &Issuer{secret: secret, ttl: ttl, now: time.Now}
}

// Issue signs a new access token for the supplied user identity.
func (i *Issuer) Issue(userID, email string) (string, time.Time, error) {
	now := i.now().UTC()
	expires := now.Add(i.ttl)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"sub":     userID,
		"iss":     "mergemarket-auth",
		"user_id": userID,
		"email":   email,
		"iat":     now.Unix(),
		"exp":     expires.Unix(),
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token: marshal header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token: marshal claims: %w", err)
	}
	unsigned := strings.Join([]string{
		base64.RawURLEncoding.EncodeToString(headerJSON),
		base64.RawURLEncoding.EncodeToString(claimsJSON),
	}, ".")
	mac := hmac.New(sha256.New, i.secret)
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		return "", time.Time{}, fmt.Errorf("token: sign input: %w", err)
	}
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsigned + "." + signature, expires, nil
}
