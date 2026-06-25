package main

import "time"

// User represents a registered user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}


const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusClosed     = "closed"
)


type Ticket struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}


func validNextStatus(current, next string) bool {
	switch current {
	case StatusOpen:
		return next == StatusInProgress
	case StatusInProgress:
		return next == StatusClosed
	case StatusClosed:
		return false
	default:
		return false
	}
}

func isValidStatus(s string) bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusClosed
}
