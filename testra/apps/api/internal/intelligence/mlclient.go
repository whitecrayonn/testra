package intelligence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MLClient is defined in ports.go.

// NewMLClient returns a real HTTP ML client when baseURL is provided, otherwise
// a transparent local heuristic client that requires no external service.
func NewMLClient(baseURL string) MLClient {
	if baseURL == "" {
		return &localMLClient{}
	}
	return &httpMLClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type httpMLClient struct {
	baseURL string
	client  *http.Client
}

func (c *httpMLClient) PredictFlaky(ctx context.Context, input PredictionInput) (PredictionResult, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return PredictionResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict-flaky", bytes.NewReader(payload))
	if err != nil {
		return PredictionResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return PredictionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return PredictionResult{}, fmt.Errorf("ml service returned %d: %s", resp.StatusCode, string(body))
	}

	var result PredictionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PredictionResult{}, err
	}
	return result, nil
}

func (c *httpMLClient) ClassifyFailure(ctx context.Context, errorMessage string, stackTrace string) (ClassificationResult, error) {
	payload, err := json.Marshal(map[string]string{
		"error_message": errorMessage,
		"stack_trace":   stackTrace,
	})
	if err != nil {
		return ClassificationResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/classify-failure", bytes.NewReader(payload))
	if err != nil {
		return ClassificationResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return ClassificationResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ClassificationResult{}, fmt.Errorf("ml service returned %d: %s", resp.StatusCode, string(body))
	}

	var result ClassificationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ClassificationResult{}, err
	}
	return result, nil
}

// localMLClient provides deterministic, explainable heuristics without external calls.
// It satisfies the "no external LLM" rule and keeps tests self-contained.
type localMLClient struct{}

func (c *localMLClient) PredictFlaky(ctx context.Context, input PredictionInput) (PredictionResult, error) {
	n := len(input.History)
	if n == 0 {
		return PredictionResult{FlakinessScore: 0, Confidence: 0.5, Explanation: "no history available"}, nil
	}

	failures := 0
	passes := 0
	transitions := 0
	var prev string
	var totalDuration int64
	for _, h := range input.History {
		switch strings.ToLower(h.Status) {
		case "failed":
			failures++
		case "passed":
			passes++
		}
		s := strings.ToLower(h.Status)
		if prev != "" && s != prev {
			transitions++
		}
		prev = s
		totalDuration += h.DurationMs
	}

	flakiness := float64(transitions) / float64(n)
	if n > 0 {
		failRatio := float64(failures) / float64(n)
		flakiness = (flakiness + failRatio) / 2
	}
	if flakiness > 1 {
		flakiness = 1
	}

	avgDur := float64(totalDuration) / float64(n)
	explanation := fmt.Sprintf("%d runs, %d transitions, %d failures, avg duration %.0fms", n, transitions, failures, avgDur)
	confidence := minFloat64(0.9, 0.5+float64(n)*0.05)
	if confidence > 1 {
		confidence = 1
	}

	return PredictionResult{
		FlakinessScore: flakiness,
		Confidence:     confidence,
		Explanation:    explanation,
	}, nil
}

func (c *localMLClient) ClassifyFailure(ctx context.Context, errorMessage string, stackTrace string) (ClassificationResult, error) {
	text := strings.ToLower(errorMessage + " " + stackTrace)
	label, explanation := classifyByKeywords(text)
	return ClassificationResult{
		Label:       label,
		Confidence:  0.75,
		Explanation: explanation,
	}, nil
}

func classifyByKeywords(text string) (string, string) {
	switch {
	case containsAny(text, "timeout", "timed out", "deadline exceeded"):
		return "timeout", "failure text mentions a timeout or deadline"
	case containsAny(text, "network", "connection", "econnrefused", "socket"):
		return "network", "failure text mentions network/connection issues"
	case containsAny(text, "assert", "assertion", "expected", "actual"):
		return "assertion", "failure text contains an assertion mismatch"
	case containsAny(text, "selector", "element", "dom", "not found"):
		return "ui_element", "failure text indicates a missing UI element or selector"
	case containsAny(text, "permission", "unauthorized", "forbidden"):
		return "authorization", "failure text indicates an authorization/permission problem"
	default:
		return "unknown", "failure does not match a known keyword class"
	}
}

func containsAny(text string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
