package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthLive(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "alive", Service: "averna-api"})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type %q", got)
	}
}
