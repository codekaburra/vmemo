package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatSendsCorrectRequest(t *testing.T) {
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("unmarshal request: %v", err)
		}
		if err := json.NewEncoder(w).Encode(chatResponse{
			Message: chatMessage{Role: "assistant", Content: "test reply"},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	origURL := ollamaURL
	ollamaURL = srv.URL
	defer func() { ollamaURL = origURL }()

	reply, err := Chat("mistral:7b", "be helpful", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "test reply" {
		t.Errorf("reply = %q, want %q", reply, "test reply")
	}
	if got.Model != "mistral:7b" {
		t.Errorf("model = %q, want %q", got.Model, "mistral:7b")
	}
	if len(got.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(got.Messages))
	}
	if got.Messages[0].Role != "system" || got.Messages[0].Content != "be helpful" {
		t.Errorf("system message = %+v", got.Messages[0])
	}
	if got.Messages[1].Role != "user" || got.Messages[1].Content != "hello" {
		t.Errorf("user message = %+v", got.Messages[1])
	}
	if got.Stream != false {
		t.Error("stream should be false")
	}
}

func TestChatHandlesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"model not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	origURL := ollamaURL
	ollamaURL = srv.URL
	defer func() { ollamaURL = origURL }()

	_, err := Chat("badmodel", "sys", "hi")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
