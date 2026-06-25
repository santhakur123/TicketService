package main

import (
	"sync"
	"time"
)



type Store struct {
	mu sync.RWMutex

	usersByID    map[string]*User
	usersByEmail map[string]*User

	tickets map[string]*Ticket
}

func NewStore() *Store {
	return &Store{
		usersByID:    make(map[string]*User),
		usersByEmail: make(map[string]*User),
		tickets:      make(map[string]*Ticket),
	}
}

func (s *Store) CreateUser(email, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return nil, ErrEmailTaken
	}

	u := &User{
		ID:           newID(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.usersByID[u.ID] = u
	s.usersByEmail[u.Email] = u
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByEmail[email]
	return u, ok
}

func (s *Store) GetUserByID(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByID[id]
	return u, ok
}

func (s *Store) CreateTicket(userID, title, description string) *Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	t := &Ticket{
		ID:          newID(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.tickets[t.ID] = t
	return t
}

func (s *Store) ListTicketsByUser(userID string) []*Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Ticket, 0)
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) GetTicket(id string) (*Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	return t, ok
}

func (s *Store) UpdateTicketStatus(id, newStatus string) (*Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}
	if !validNextStatus(t.Status, newStatus) {
		return nil, ErrInvalidTransition
	}
	t.Status = newStatus
	t.UpdatedAt = time.Now()
	return t, nil
}
