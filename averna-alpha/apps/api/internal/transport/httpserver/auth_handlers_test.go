package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"a@example.com","unexpected":true}`))
	var payload struct {
		Email string `json:"email"`
	}
	if decodeJSON(recorder, request, &payload) {
		t.Fatal("expected unknown field to be rejected")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestDecodeJSONRejectsMalformedBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":`))
	var payload struct {
		Email string `json:"email"`
	}
	if decodeJSON(recorder, request, &payload) {
		t.Fatal("expected malformed JSON to be rejected")
	}
}

func TestDecodeJSONRejectsTrailingValue(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"a@example.com"} {"email":"b@example.com"}`))
	var payload struct {
		Email string `json:"email"`
	}
	if decodeJSON(recorder, request, &payload) {
		t.Fatal("expected trailing JSON value to be rejected")
	}
}
