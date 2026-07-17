package integrationhub

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

// dispatch sends an integration payload to the configured external service.
// It returns the external reference ID (if any) and an error.
func dispatch(ctx context.Context, integration *Integration, payload map[string]interface{}) (string, error) {
	adapter := adapterFor(integration.Type)
	return adapter.Send(ctx, integration, payload)
}

func adapterFor(t IntegrationType) Adapter {
	switch t {
	case TypeJira:
		return &jiraAdapter{}
	case TypeGitHub:
		return &githubAdapter{}
	case TypeGitLab:
		return &gitlabAdapter{}
	case TypeSlack:
		return &slackAdapter{}
	default:
		return &webhookAdapter{}
	}
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

type jiraAdapter struct{}

func (a *jiraAdapter) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	baseURL := strings.TrimRight(i.Config["url"], "/")
	token := i.Config["token"]
	projectKey := i.Config["project_key"]
	if baseURL == "" || token == "" || projectKey == "" {
		return "", fmt.Errorf("jira integration missing url, token or project_key")
	}

	summary, _ := payload["summary"].(string)
	if summary == "" {
		summary = "Testra event"
	}
	description, _ := payload["description"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"fields": map[string]interface{}{
			"project":   map[string]string{"key": projectKey},
			"summary":   summary,
			"description": description,
			"issuetype": map[string]string{"name": "Bug"},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/rest/api/3/issue", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(i.Config["username"], token))

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("jira returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
		Key string `json:"key"`
	}
	_ = json.Unmarshal(respBody, &result)
	return result.Key, nil
}

type githubAdapter struct{}

func (a *githubAdapter) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	baseURL := strings.TrimRight(i.Config["url"], "/")
	token := i.Config["token"]
	owner := i.Config["owner"]
	repo := i.Config["repo"]
	if baseURL == "" || token == "" || owner == "" || repo == "" {
		return "", fmt.Errorf("github integration missing url, token, owner or repo")
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	bodyText, _ := payload["body"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"title": title,
		"body":  bodyText,
	})

	u := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Number int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("#%d", result.Number), nil
}

type gitlabAdapter struct{}

func (a *gitlabAdapter) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	baseURL := strings.TrimRight(i.Config["url"], "/")
	token := i.Config["token"]
	projectID := i.Config["project_id"]
	if baseURL == "" || token == "" || projectID == "" {
		return "", fmt.Errorf("gitlab integration missing url, token or project_id")
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	description, _ := payload["description"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"title":       title,
		"description": description,
	})

	u := fmt.Sprintf("%s/api/v4/projects/%s/issues", baseURL, projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		IID int `json:"iid"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("!%d", result.IID), nil
}

type slackAdapter struct{}

func (a *slackAdapter) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	url := i.Config["url"]
	if url == "" {
		return "", fmt.Errorf("slack integration missing webhook url")
	}

	text, _ := payload["text"].(string)
	if text == "" {
		text = "Testra event"
	}

	body, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("slack returned %d: %s", resp.StatusCode, string(respBody))
	}
	return "ok", nil
}

type webhookAdapter struct{}

func (a *webhookAdapter) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	url := i.Config["url"]
	if url == "" {
		return "", fmt.Errorf("webhook integration missing url")
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := i.Config["secret"]; secret != "" {
		req.Header.Set("X-Webhook-Secret", secret)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func basicAuth(username, password string) string {
	if username == "" {
		username = ""
	}
	return base64Encode(username + ":" + password)
}

func base64Encode(s string) string {
	// Minimal base64 to avoid importing encoding/base64 for a single usage.
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var buf bytes.Buffer
	var pad int
	for i := 0; i < len(s); i += 3 {
		b := []byte(s[i:])
		if len(b) > 3 {
			b = b[:3]
		}
		var val uint32
		for j := 0; j < len(b); j++ {
			val = val<<8 | uint32(b[j])
		}
		pad = 3 - len(b)
		for j := 0; j < 4-pad; j++ {
			idx := (val >> (18 - j*6)) & 0x3F
			buf.WriteByte(alphabet[idx])
		}
		for j := 0; j < pad; j++ {
			buf.WriteByte('=')
		}
	}
	return buf.String()
}
