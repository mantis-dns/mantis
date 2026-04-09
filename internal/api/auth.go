package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mantis-dns/mantis/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	sessions domain.SessionRepository
	settings domain.SettingsRepository
}

type setupRequest struct {
	Password string `json:"password"`
}

type loginRequest struct {
	Password string `json:"password"`
}

// Setup creates the admin account (first-run only).
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	setupDone, _ := h.settings.Get(r.Context(), "auth.setup_completed")
	if setupDone == "true" {
		Error(w, "FORBIDDEN", "setup already completed", http.StatusForbidden)
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		Error(w, "VALIDATION_ERROR", "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to hash password", http.StatusInternalServerError)
		return
	}

	h.settings.Set(r.Context(), "auth.password_hash", string(hash))
	h.settings.Set(r.Context(), "auth.setup_completed", "true")
	Success(w, map[string]string{"status": "setup complete"})
}

// Login authenticates and creates a session.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}

	hash, err := h.settings.Get(r.Context(), "auth.password_hash")
	if err != nil || hash == "" {
		Error(w, "UNAUTHORIZED", "run setup first", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		Error(w, "UNAUTHORIZED", "invalid password", http.StatusUnauthorized)
		return
	}

	token := generateToken()
	session := &domain.Session{
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IP:        clientIP(r),
	}
	h.sessions.CreateSession(r.Context(), session)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	Success(w, map[string]string{"status": "authenticated"})
}

// Logout invalidates the session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.sessions.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	Success(w, map[string]string{"status": "logged out"})
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

func generateAPIKeyToken() (token, hash string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	token = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(h[:])
	return token, hash
}

func init() {
	_ = uuid.New // ensure uuid is used
}
