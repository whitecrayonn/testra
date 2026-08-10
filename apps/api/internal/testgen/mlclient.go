package testgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// FileGenerationClient calls the ML service's LLM-backed test case
// generation endpoint. This is the one intentional exception to this
// package's "no AI/ML" rule — see the package doc comment in domain.go and
// docs/BIBLICAL_TESTRA.md's "No External LLM" principle.
type FileGenerationClient interface {
	GenerateFromFile(ctx context.Context, filename string, content []byte, promptContext string) (*fileGenerationResponse, error)
}

// MLServiceError preserves the ML service's HTTP status and message so the
// handler can surface an accurate, specific error (e.g. "LLM not
// configured" vs "rate limited") instead of a generic failure.
type MLServiceError struct {
	StatusCode int
	Message    string
}

func (e *MLServiceError) Error() string { return e.Message }

// NewFileGenerationClient returns a real HTTP client when baseURL is
// configured, otherwise a client that always fails loudly with "not
// configured" rather than silently no-op'ing or fabricating test cases —
// unlike intelligence.NewMLClient, there is no deterministic local fallback
// for freeform LLM generation.
func NewFileGenerationClient(baseURL, apiKey string) FileGenerationClient {
	if strings.TrimSpace(baseURL) == "" {
		return &unconfiguredFileGenerationClient{}
	}
	return &httpFileGenerationClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type unconfiguredFileGenerationClient struct{}

func (c *unconfiguredFileGenerationClient) GenerateFromFile(_ context.Context, _ string, _ []byte, _ string) (*fileGenerationResponse, error) {
	return nil, &MLServiceError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "AI generation is not configured on this server (ML_SERVICE_URL is unset).",
	}
}

type httpFileGenerationClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type generatedStepResponse struct {
	Action   string `json:"action"`
	Expected string `json:"expected"`
	TestData string `json:"test_data"`
}

type generatedCaseFromFile struct {
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	Preconditions string                  `json:"preconditions"`
	Priority      string                  `json:"priority"`
	Tags          []string                `json:"tags"`
	Steps         []generatedStepResponse `json:"steps"`
}

type skippedRowResponse struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type fileGenerationResponse struct {
	Cases       []generatedCaseFromFile `json:"cases"`
	SkippedRows []skippedRowResponse    `json:"skipped_rows"`
	RowCount    int                     `json:"row_count"`
}

func (c *httpFileGenerationClient) GenerateFromFile(ctx context.Context, filename string, content []byte, promptContext string) (*fileGenerationResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	if promptContext != "" {
		if err := writer.WriteField("context", promptContext); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/generate-test-cases-from-file", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &MLServiceError{StatusCode: http.StatusBadGateway, Message: fmt.Sprintf("could not reach the ML service: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &MLServiceError{StatusCode: mapUpstreamStatus(resp.StatusCode), Message: extractDetail(respBody)}
	}

	var result fileGenerationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &MLServiceError{StatusCode: http.StatusBadGateway, Message: "ML service returned an unreadable response"}
	}
	return &result, nil
}

// mapUpstreamStatus passes through the status codes the ML service uses
// deliberately (400 bad input, 503 not configured, 502 upstream/LLM
// failure); anything else collapses to 502 so callers see "upstream
// failed" rather than an unexpected status they'd have to special-case.
func mapUpstreamStatus(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusServiceUnavailable, http.StatusBadGateway:
		return status
	default:
		return http.StatusBadGateway
	}
}

func extractDetail(body []byte) string {
	var parsed struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Detail != "" {
		return parsed.Detail
	}
	if len(body) == 0 {
		return "ML service request failed"
	}
	return string(body)
}
