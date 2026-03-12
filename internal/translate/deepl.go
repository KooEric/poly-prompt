package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultBaseURL = "https://api-free.deepl.com"

var ErrMissingAPIKey = errors.New("DEEPL_API_KEY is not set")

type Translator interface {
	Translate(ctx context.Context, text string) (string, error)
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient HTTPDoer
}

type DeepLClient struct {
	apiKey     string
	baseURL    string
	httpClient HTTPDoer
}

type requestBody struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
}

type responseBody struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

func NewDeepLClient(opts ClientOptions) *DeepLClient {
	baseURL := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &DeepLClient{
		apiKey:     strings.TrimSpace(opts.APIKey),
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *DeepLClient) Translate(ctx context.Context, text string) (string, error) {
	if c.apiKey == "" {
		return "", ErrMissingAPIKey
	}

	if strings.TrimSpace(text) == "" {
		return "", errors.New("input text is empty")
	}

	payload, err := json.Marshal(requestBody{
		Text:       []string{text},
		TargetLang: "EN-US",
	})
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v2/translate", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("translate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("translate request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded responseBody
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode translation response: %w", err)
	}

	if len(decoded.Translations) == 0 || strings.TrimSpace(decoded.Translations[0].Text) == "" {
		return "", errors.New("translate response did not include any translations")
	}

	return decoded.Translations[0].Text, nil
}
