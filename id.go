package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)


func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var (
	ErrEmailTaken        = errors.New("email already registered")
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrInvalidTransition = errors.New("invalid status transition")
)
