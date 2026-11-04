package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/avernus666/averna/apps/api/internal/organizations"
)

type organizationHandler struct {
	repository *organizations.Repository
	logger     *slog.Logger
}

func newOrganizationHandler(repository *organizations.Repository, logger *slog.Logger) *organizationHandler {
	return &organizationHandler{repository: repository, logger: logger}
}

func (h *organizationHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return
	}
	items, err := h.repository.ListForUser(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("list organizations failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not load organizations")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
