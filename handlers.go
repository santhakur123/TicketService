package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type App struct {
	store     *Store
	jwtSecret []byte
}

func NewApp(store *Store, jwtSecret []byte) *App {
	return &App{store: store, jwtSecret: jwtSecret}
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSONBody(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(dst)
}

type contextKey string

const userIDKey contextKey = "userID"

// authMiddleware validates the Bearer token and injects the user ID into
// the request context for downstream handlers.
func (a *App) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := parseAndVerifyToken(a.jwtSecret, tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		if _, ok := a.store.GetUserByID(claims.UserID); !ok {
			writeError(w, http.StatusUnauthorized, "user no longer exists")
			return
		}

		ctx := withUserID(r.Context(), claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

// ---------- handlers ----------

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user, err := a.store.CreateUser(req.Email, hash)
	if err != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":    user.ID,
		"email": user.Email,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, ok := a.store.GetUserByEmail(req.Email)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	valid, err := verifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := generateToken(a.jwtSecret, user.ID, user.Email, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (a *App) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createTicketRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	ticket := a.store.CreateTicket(userID, req.Title, req.Description)
	writeJSON(w, http.StatusCreated, ticket)
}

func (a *App) handleListTickets(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tickets := a.store.ListTicketsByUser(userID)
	writeJSON(w, http.StatusOK, tickets)
}

func (a *App) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	ticket, found := a.store.GetTicket(id)
	if !found {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		// Do not leak existence of other users' tickets.
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (a *App) handleUpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := r.PathValue("id")
	ticket, found := a.store.GetTicket(id)
	if !found {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	var req updateStatusRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !isValidStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "status must be one of: open, in_progress, closed")
		return
	}

	updated, err := a.store.UpdateTicketStatus(id, req.Status)
	if err != nil {
		switch err {
		case ErrInvalidTransition:
			writeError(w, http.StatusBadRequest, "invalid status transition: "+ticket.Status+" -> "+req.Status)
		case ErrTicketNotFound:
			writeError(w, http.StatusNotFound, "ticket not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update ticket")
		}
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
