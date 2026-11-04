package telegram

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookRejectsMissingSecret(t *testing.T) {
	h := NewWebhook(Config{BotToken: "token", WebhookSecret: "expected-secret", MiniAppURL: "https://mini.example"})
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebhookDisabledWhenIncomplete(t *testing.T) {
	h := NewWebhook(Config{BotToken: "token", WebhookSecret: "secret"})
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestWebhookRejectsWrongMethod(t *testing.T) {
	h := NewWebhook(Config{BotToken: "token", WebhookSecret: "expected-secret", MiniAppURL: "https://mini.example"})
	req := httptest.NewRequest(http.MethodGet, "/telegram/webhook", nil)
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "expected-secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestStoreURLAddsValidatedSlug(t *testing.T) {
	got := storeURL("https://mini.example/app?theme=dark", "north-store")
	if got != "https://mini.example/app?store=north-store&theme=dark" {
		t.Fatalf("unexpected Mini App URL: %s", got)
	}
}

func TestValidStoreSlug(t *testing.T) {
	for _, value := range []string{"shop", "north-store", "store2026"} {
		if !validStoreSlug(value) {
			t.Errorf("expected %q to be valid", value)
		}
	}
	for _, value := range []string{"", "ab", "-shop", "shop-", "SHOP", "shop_1", "shop?x=1"} {
		if validStoreSlug(value) {
			t.Errorf("expected %q to be invalid", value)
		}
	}
}
