package translate

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDeepLClientTranslateSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/translate" {
			t.Fatalf("request path = %q, want %q", r.URL.Path, "/v2/translate")
		}
		if got := r.Header.Get("Authorization"); got != "DeepL-Auth-Key test-key" {
			t.Fatalf("Authorization header = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type header = %q", got)
		}

		var payload requestBody
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(payload.Text) != 1 || payload.Text[0] != "안녕하세요" {
			t.Fatalf("payload.Text = %#v", payload.Text)
		}
		if payload.TargetLang != "EN-US" {
			t.Fatalf("payload.TargetLang = %q", payload.TargetLang)
		}

		_, _ = w.Write([]byte(`{"translations":[{"text":"Hello"}]}`))
	}))
	defer server.Close()

	client := NewDeepLClient(ClientOptions{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	})

	got, err := client.Translate(context.Background(), "안녕하세요")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if got != "Hello" {
		t.Fatalf("Translate() = %q, want %q", got, "Hello")
	}
}

func TestDeepLClientTranslateErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		apiKey    string
		handler   http.HandlerFunc
		wantError string
	}{
		{
			name:      "missing api key",
			apiKey:    "",
			wantError: ErrMissingAPIKey.Error(),
		},
		{
			name:   "non-200 response",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "bad request", http.StatusBadRequest)
			},
			wantError: "status 400",
		},
		{
			name:   "malformed json",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"translations":`))
			},
			wantError: "decode translation response",
		},
		{
			name:   "empty translations",
			apiKey: "test-key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"translations":[]}`))
			},
			wantError: "did not include any translations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			baseURL := "https://example.invalid"
			httpClient := http.DefaultClient
			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				defer server.Close()
				baseURL = server.URL
				httpClient = server.Client()
			}

			client := NewDeepLClient(ClientOptions{
				APIKey:     tt.apiKey,
				BaseURL:    baseURL,
				HTTPClient: httpClient,
			})

			_, err := client.Translate(context.Background(), "안녕하세요")
			if err == nil {
				t.Fatal("Translate() expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("Translate() error = %v, want substring %q", err, tt.wantError)
			}
		})
	}
}

func TestDeepLClientTranslateTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"translations":[{"text":"Hello"}]}`))
	}))
	defer server.Close()

	client := NewDeepLClient(ClientOptions{
		APIKey:  "test-key",
		BaseURL: server.URL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Millisecond,
		},
	})

	_, err := client.Translate(context.Background(), "안녕하세요")
	if err == nil {
		t.Fatal("Translate() expected a timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "Client.Timeout") {
		t.Fatalf("Translate() error = %v, want timeout-related error", err)
	}
}
