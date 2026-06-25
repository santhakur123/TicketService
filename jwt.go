package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)



var ErrInvalidToken = errors.New("invalid or expired token")

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func generateToken(secret []byte, userID, email string, ttl time.Duration) (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwtClaims{
		UserID: userID,
		Email:  email,
		Iat:    now.Unix(),
		Exp:    now.Add(ttl).Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := base64URLEncode(headerJSON)
	claimsPart := base64URLEncode(claimsJSON)
	signingInput := headerPart + "." + claimsPart

	sig := signHMAC(secret, signingInput)
	sigPart := base64URLEncode(sig)

	return signingInput + "." + sigPart, nil
}

func signHMAC(secret []byte, data string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func parseAndVerifyToken(secret []byte, token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	headerPart, claimsPart, sigPart := parts[0], parts[1], parts[2]

	expectedSig := signHMAC(secret, headerPart+"."+claimsPart)
	actualSig, err := base64URLDecode(sigPart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := base64URLDecode(claimsPart)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
