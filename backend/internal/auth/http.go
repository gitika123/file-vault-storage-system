package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/balkanid/file-vault/internal/auth/policy"
	"github.com/balkanid/file-vault/internal/auth/session"
)

type principalContextKey struct{}
type csrfHashContextKey struct{}

type HTTPHandler struct {
	Service      Service
	CookieSecure bool
	Now          func() time.Time
	RateLimiter  interface{ Allow(string) bool }
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

func (h HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&input); err != nil || strings.TrimSpace(input.Email) == "" || input.Password == "" {
		writeAuthError(w, http.StatusBadRequest, "INVALID_INPUT", "Email and password are required.")
		return
	}
	user, newSession, err := h.Service.Authenticate(r.Context(), input.Email, input.Password, h.clock())
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid email or password.")
		case errors.Is(err, ErrAccountDisabled):
			writeAuthError(w, http.StatusForbidden, "FORBIDDEN", "This account is disabled.")
		default:
			writeAuthError(w, http.StatusInternalServerError, "INTERNAL", "Unable to sign in.")
		}
		return
	}
	session.SetCookie(w, newSession.RawToken, newSession.ExpiresAt, h.CookieSecure)
	session.SetCSRFCookie(w, newSession.CSRFToken, newSession.ExpiresAt, h.CookieSecure)
	writeJSON(w, http.StatusOK, userResponseFrom(user))
}

func (h HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input registerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&input); err != nil {
		writeAuthError(w, http.StatusBadRequest, "INVALID_INPUT", "Enter your name, email, and password.")
		return
	}
	user, newSession, err := h.Service.Register(r.Context(), input.Email, input.DisplayName, input.Password, h.clock())
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeAuthError(w, http.StatusConflict, "EMAIL_TAKEN", "An account with this email already exists.")
			return
		}
		if strings.Contains(err.Error(), "valid email") || strings.Contains(err.Error(), "password must") {
			writeAuthError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
			return
		}
		writeAuthError(w, http.StatusInternalServerError, "INTERNAL", "Unable to create your account.")
		return
	}
	session.SetCookie(w, newSession.RawToken, newSession.ExpiresAt, h.CookieSecure)
	session.SetCSRFCookie(w, newSession.CSRFToken, newSession.ExpiresAt, h.CookieSecure)
	writeJSON(w, http.StatusCreated, userResponseFrom(user))
}

func (h HTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if token := sessionToken(r); token != "" {
		if err := h.Service.Revoke(r.Context(), token); err != nil {
			writeAuthError(w, http.StatusInternalServerError, "INTERNAL", "Unable to sign out.")
			return
		}
	}
	session.ClearCookie(w, h.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h HTTPHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.Service.UserForToken(r.Context(), sessionToken(r), h.clock())
	if err != nil {
		writeAuthError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	writeJSON(w, http.StatusOK, userResponseFrom(user))
}

func (h HTTPHandler) CSRF(w http.ResponseWriter, r *http.Request) {
	if _, ok := csrfHashFromContext(r.Context()); !ok {
		writeAuthError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	cookie, err := r.Cookie(session.CSRFCookie)
	if err != nil || cookie.Value == "" {
		writeAuthError(w, http.StatusForbidden, "FORBIDDEN", "CSRF token is unavailable; sign in again.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"csrfToken": cookie.Value})
}

func (h HTTPHandler) RequireSession(requireCSRF bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, csrfHash, err := h.Service.Principal(r.Context(), sessionToken(r), h.clock())
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		ctx = context.WithValue(ctx, csrfHashContextKey{}, csrfHash)
		// SSE connections are long-lived subscriptions, not request bursts.
		// Counting them would make the dashboard throttle its own refreshes.
		if h.RateLimiter != nil && r.URL.Path != "/api/events/downloads" && !h.RateLimiter.Allow(principal.UserID) {
			w.Header().Set("Retry-After", "1")
			writeAuthError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests; try again shortly.")
			return
		}
		if requireCSRF {
			if err := session.ValidateCSRF(r, csrfHash); err != nil {
				writeAuthError(w, http.StatusForbidden, "FORBIDDEN", "Invalid CSRF token.")
				return
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) (*policy.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(policy.Principal)
	if !ok {
		return nil, false
	}
	return &principal, true
}

func csrfHashFromContext(ctx context.Context) ([32]byte, bool) {
	hash, ok := ctx.Value(csrfHashContextKey{}).([32]byte)
	return hash, ok
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(session.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h HTTPHandler) clock() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

func userResponseFrom(user User) userResponse {
	return userResponse{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: string(user.Role)}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
