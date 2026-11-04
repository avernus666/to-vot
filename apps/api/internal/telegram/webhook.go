package telegram

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BotToken      string
	WebhookSecret string
	MiniAppURL    string
}

type Webhook struct {
	cfg    Config
	client *http.Client
}

type update struct {
	Message *message `json:"message"`
}

type message struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Text string `json:"text"`
}

type replyMarkup struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

type inlineButton struct {
	Text   string  `json:"text"`
	WebApp *webApp `json:"web_app,omitempty"`
	URL    string  `json:"url,omitempty"`
}

type webApp struct {
	URL string `json:"url"`
}

func NewWebhook(cfg Config) *Webhook {
	return &Webhook{cfg: cfg, client: &http.Client{Timeout: 8 * time.Second}}
}

func (h *Webhook) Enabled() bool {
	return h.cfg.BotToken != "" && h.cfg.WebhookSecret != "" && h.cfg.MiniAppURL != ""
}

func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.Enabled() {
		http.Error(w, "telegram integration is not configured", http.StatusServiceUnavailable)
		return
	}
	provided := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if len(provided) != len(h.cfg.WebhookSecret) || subtle.ConstantTimeCompare([]byte(provided), []byte(h.cfg.WebhookSecret)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var incoming update
	if err := json.Unmarshal(body, &incoming); err != nil {
		http.Error(w, "invalid update", http.StatusBadRequest)
		return
	}
	if incoming.Message == nil || incoming.Message.Chat.ID == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	text := strings.TrimSpace(incoming.Message.Text)
	if text == "/start" || strings.HasPrefix(text, "/start ") || text == "/catalog" || text == "Каталог" {
		miniAppURL := h.cfg.MiniAppURL
		if strings.HasPrefix(text, "/start ") {
			payload := strings.TrimSpace(strings.TrimPrefix(text, "/start "))
			if validStoreSlug(payload) {
				miniAppURL = storeURL(h.cfg.MiniAppURL, payload)
			}
		}
		if err := h.sendWelcome(r.Context(), incoming.Message.Chat.ID, miniAppURL); err != nil {
			http.Error(w, "telegram delivery failed", http.StatusBadGateway)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Webhook) sendWelcome(ctx context.Context, chatID int64, miniAppURL string) error {
	payload := map[string]any{
		"chat_id":      chatID,
		"text":         "Добро пожаловать в AVERNA! Откройте приложение, чтобы посмотреть каталог или управлять своим бизнесом.",
		"reply_markup": replyMarkup{InlineKeyboard: [][]inlineButton{{{Text: "Открыть AVERNA", WebApp: &webApp{URL: miniAppURL}}}}},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode Telegram message: %w", err)
	}
	endpoint := "https://api.telegram.org/bot" + h.cfg.BotToken + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("send Telegram message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Telegram API returned status %d", resp.StatusCode)
	}
	return nil
}

func validStoreSlug(value string) bool {
	if len(value) < 3 || len(value) > 48 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

func storeURL(base, slug string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := parsed.Query()
	query.Set("store", slug)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
