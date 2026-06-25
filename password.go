package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
)



const hashRounds = 100000

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := derive(password, salt)
	return fmt.Sprintf("%s:%s", hex.EncodeToString(salt), hex.EncodeToString(hash)), nil
}

func verifyPassword(password, stored string) (bool, error) {
	parts := splitOnce(stored, ':')
	if parts == nil {
		return false, errors.New("invalid stored hash format")
	}
	saltHex, hashHex := parts[0], parts[1]

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, err
	}
	expected, err := hex.DecodeString(hashHex)
	if err != nil {
		return false, err
	}
	actual := derive(password, salt)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func derive(password string, salt []byte) []byte {
	data := append(append([]byte{}, salt...), []byte(password)...)
	sum := sha256.Sum256(data)
	for i := 0; i < hashRounds; i++ {
		sum = sha256.Sum256(sum[:])
	}
	return sum[:]
}

func splitOnce(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
