package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/avernus666/averna/apps/api/internal/auth"
)

type authHandler struct {
	service *auth.Service
}

func newAuthHandler(service *auth.Service) *authHandler {
	return &authHandler{service: service}
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		OrgName     string `json:"organization_name"`
		OrgSlug     string `json:"organization_slug"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, tokens, err := h.service.Register(r.Context(), auth.RegisterInput{
		Email: input.Email, Password: input.Password, DisplayName: input.DisplayName,
		OrgName: input.OrgName, OrgSlug: input.OrgSlug,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email_taken", "An account with this email already exists")
		default:
			writeError(w, http.StatusBadRequest, "invalid_registration", err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, tokens, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect")
			return
		}
		writeError(w, http.StatusInternalServerError, "login_failed", "Could not complete sign-in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": tokens})
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	tokens, err := h.service.Refresh(r.Context(), input.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefresh) {
			writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "The refresh token is invalid, expired, or already rotated")
			return
		}
		writeError(w, http.StatusInternalServerError, "refresh_failed", "Could not refresh session")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.service.Logout(r.Context(), input.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "logout_failed", "Could not revoke session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON with supported fields")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must contain a single JSON value")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": strings.TrimSpace(message)},
	})
}
