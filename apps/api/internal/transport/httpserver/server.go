package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/avernus666/averna/apps/api/internal/auth"
	"github.com/avernus666/averna/apps/api/internal/organizations"
	"github.com/avernus666/averna/apps/api/internal/telegram"
	"github.com/jackc/pgx/v5/pgxpool"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type TelegramConfig struct {
	BotToken      string
	WebhookSecret string
	MiniAppURL    string
}

func New(db *pgxpool.Pool, logger *slog.Logger, telegramCfg TelegramConfig) http.Handler {
	mux := http.NewServeMux()
	authService := auth.NewService(db)
	authHandlers := newAuthHandler(authService)
	orgHandlers := newOrganizationHandler(organizations.NewRepository(db), logger)
	commerce := newCommerceHandler(db)
	telegramWebhook := telegram.NewWebhook(telegram.Config{BotToken: telegramCfg.BotToken, WebhookSecret: telegramCfg.WebhookSecret, MiniAppURL: telegramCfg.MiniAppURL})

	mux.Handle("POST /telegram/webhook", telegramWebhook)
	mux.HandleFunc("GET /v1/store/{slug}/products", commerce.publicProducts)
	mux.HandleFunc("POST /v1/store/{slug}/orders", commerce.createOrder)
	mux.Handle("GET /v1/organizations/{organizationID}/products", requireAuth(authService, http.HandlerFunc(commerce.organizationProducts)))
	mux.Handle("POST /v1/organizations/{organizationID}/products", requireAuth(authService, http.HandlerFunc(commerce.organizationProducts)))
	mux.Handle("PATCH /v1/organizations/{organizationID}/products/{productID}", requireAuth(authService, http.HandlerFunc(commerce.organizationProducts)))
	mux.Handle("GET /v1/organizations/{organizationID}/orders", requireAuth(authService, http.HandlerFunc(commerce.organizationOrders)))
	mux.Handle("PATCH /v1/organizations/{organizationID}/orders/{orderID}", requireAuth(authService, http.HandlerFunc(commerce.updateOrderStatus)))

	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "alive", Service: "averna-api"})
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			logger.Warn("readiness probe failed", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "not_ready", Service: "averna-api"})
			return
		}
		writeJSON(w, http.StatusOK, healthResponse{Status: "ready", Service: "averna-api"})
	})

	mux.HandleFunc("POST /v1/auth/register", authHandlers.register)
	mux.HandleFunc("POST /v1/auth/login", authHandlers.login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandlers.refresh)
	mux.HandleFunc("POST /v1/auth/logout", authHandlers.logout)
	mux.Handle("GET /v1/me", requireAuth(authService, http.HandlerFunc(authHandlers.me)))
	mux.Handle("GET /v1/organizations", requireAuth(authService, http.HandlerFunc(orgHandlers.list)))

	return requestID(logging(logger, recoverer(logger, securityHeaders(cors(telegramCfg.MiniAppURL, mux)))))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
